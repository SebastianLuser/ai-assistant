package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	appdb "jarvis/db"
	"jarvis/pkg/domain"
	"jarvis/pkg/service/sqldata"

	_ "github.com/lib/pq"
)

type PGMemoryService struct {
	db *sql.DB
}

func NewPGMemoryService(dsn string) (*PGMemoryService, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreOpen, err)
	}

	if err := db.Ping(); err != nil {
		return nil, domain.Wrapf(domain.ErrStorePing, err)
	}

	if err := appdb.RunMigrations(db); err != nil {
		return nil, domain.Wrapf(domain.ErrStoreMigrate, err)
	}

	return &PGMemoryService{db: db}, nil
}

func (s *PGMemoryService) Save(content string, tags []string, embedding []float64) (int64, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return 0, domain.Wrapf(domain.ErrStoreSave, err)
	}

	embJSON, err := json.Marshal(embedding)
	if err != nil {
		return 0, domain.Wrapf(domain.ErrStoreSave, err)
	}

	var id int64
	if err := s.db.QueryRow(sqldata.SaveMemory, content, string(tagsJSON), string(embJSON)).Scan(&id); err != nil {
		return 0, domain.Wrapf(domain.ErrStoreSave, err)
	}

	return id, nil
}

func (s *PGMemoryService) Search(queryEmbedding []float64, limit int) ([]domain.Memory, error) {
	rows, err := s.db.Query(sqldata.SelectMemories)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreSearch, err)
	}
	defer rows.Close()

	type scored struct {
		mem   domain.Memory
		score float64
	}

	var results []scored
	for rows.Next() {
		var m domain.Memory
		var tagsStr, embStr string
		if err := rows.Scan(&m.ID, &m.Content, &tagsStr, &embStr, &m.CreatedAt); err != nil {
			log.Printf("pgstore: scan row: %v", err)
			continue
		}

		if err := json.Unmarshal([]byte(tagsStr), &m.Tags); err != nil {
			log.Printf("pgstore: unmarshal tags id=%d: %v", m.ID, err)
		}

		var emb []float64
		if err := json.Unmarshal([]byte(embStr), &emb); err != nil {
			log.Printf("pgstore: unmarshal embedding id=%d: %v", m.ID, err)
		}

		similarity := cosineSimilarity(queryEmbedding, emb)
		decay := timeDecay(m.CreatedAt)
		results = append(results, scored{mem: m, score: similarity * decay})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if limit > len(results) {
		limit = len(results)
	}

	memories := make([]domain.Memory, limit)
	for i := 0; i < limit; i++ {
		memories[i] = results[i].mem
		memories[i].Score = results[i].score
	}

	return memories, nil
}

func (s *PGMemoryService) SearchFTS(query string, limit int) ([]domain.Memory, error) {
	rows, err := s.db.Query(sqldata.SearchFTS, query, limit)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreFTS, err)
	}
	defer rows.Close()

	var results []domain.Memory
	for rows.Next() {
		var m domain.Memory
		var tagsStr string
		var rank float64
		if err := rows.Scan(&m.ID, &m.Content, &tagsStr, &m.CreatedAt, &rank); err != nil {
			log.Printf("pgstore: scan fts row: %v", err)
			continue
		}

		if err := json.Unmarshal([]byte(tagsStr), &m.Tags); err != nil {
			log.Printf("pgstore: unmarshal tags id=%d: %v", m.ID, err)
		}

		m.Score = rank * timeDecay(m.CreatedAt)
		results = append(results, m)
	}

	return results, nil
}

func (s *PGMemoryService) SearchHybrid(query string, queryEmbedding []float64, limit int, vecWeight, ftsWeight float64) ([]domain.Memory, error) {
	vecResults, vecErr := s.Search(queryEmbedding, limit*2)
	ftsResults, ftsErr := s.SearchFTS(query, limit*2)

	if vecErr != nil && ftsErr != nil {
		return nil, domain.Wrap(domain.ErrStoreHybrid, "both vector and fts failed")
	}

	merged := mergeSearchResults(vecResults, ftsResults, vecErr, ftsErr, vecWeight, ftsWeight)

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	if limit > len(merged) {
		limit = len(merged)
	}

	return merged[:limit], nil
}

func (s *PGMemoryService) Delete(id int64) error {
	if _, err := s.db.Exec(sqldata.DeleteMemory, id); err != nil {
		return domain.Wrapf(domain.ErrStoreDelete, err)
	}
	return nil
}

func (s *PGMemoryService) SaveConversation(sessionID, role, content string) error {
	if _, err := s.db.Exec(sqldata.SaveConversation, sessionID, role, content); err != nil {
		return domain.Wrapf(domain.ErrConversationSave, err)
	}
	return nil
}

func (s *PGMemoryService) LoadConversation(sessionID string, limit int) ([]domain.ConversationMessage, error) {
	rows, err := s.db.Query(sqldata.LoadConversation, sessionID, limit)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrConversationLoad, err)
	}
	defer rows.Close()

	var msgs []domain.ConversationMessage
	for rows.Next() {
		var m domain.ConversationMessage
		if err := rows.Scan(&m.Role, &m.Content, &m.CreatedAt); err != nil {
			log.Printf("pgstore: scan conversation row: %v", err)
			continue
		}
		msgs = append(msgs, m)
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}

func (s *PGMemoryService) ClearConversation(sessionID string) error {
	if _, err := s.db.Exec(sqldata.ClearConversation, sessionID); err != nil {
		return domain.Wrapf(domain.ErrConversationClear, err)
	}
	return nil
}

func (s *PGMemoryService) ReplaceConversation(sessionID string, msgs []domain.ConversationMessage) error {
	tx, err := s.db.Begin()
	if err != nil {
		return domain.Wrapf(domain.ErrConversationReplace, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(sqldata.ClearConversation, sessionID); err != nil {
		return domain.Wrapf(domain.ErrConversationReplace, err)
	}

	for _, m := range msgs {
		if _, err := tx.Exec(sqldata.SaveConversation, sessionID, m.Role, m.Content); err != nil {
			return domain.Wrapf(domain.ErrConversationReplace, err)
		}
	}

	return tx.Commit()
}

func (s *PGMemoryService) LogHabit(name string) error {
	if _, err := s.db.Exec(sqldata.LogHabit, name); err != nil {
		return domain.Wrapf(domain.ErrHabitLog, err)
	}
	return nil
}

func (s *PGMemoryService) GetHabitStreak(name string) (int, int, error) {
	var total int
	if err := s.db.QueryRow(sqldata.CountHabit, name).Scan(&total); err != nil {
		return 0, 0, domain.Wrapf(domain.ErrHabitQuery, err)
	}

	rows, err := s.db.Query(sqldata.HabitDates, name)
	if err != nil {
		return 0, 0, domain.Wrapf(domain.ErrHabitQuery, err)
	}
	defer rows.Close()

	streak := 0
	expected := time.Now().Truncate(24 * time.Hour)
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			break
		}
		d = d.Truncate(24 * time.Hour)
		if d.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, total, nil
}

func (s *PGMemoryService) ListHabitsToday() ([]string, error) {
	rows, err := s.db.Query(sqldata.HabitsToday)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrHabitQuery, err)
	}
	defer rows.Close()

	var habits []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		habits = append(habits, name)
	}
	return habits, nil
}

func (s *PGMemoryService) ListExpenses(from, to string) ([]domain.Expense, error) {
	rows, err := s.db.Query(sqldata.ListExpenses, from, to)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrFinanceSummary, err)
	}
	defer rows.Close()

	var expenses []domain.Expense
	for rows.Next() {
		var e domain.Expense
		if err := rows.Scan(&e.ID, &e.Date, &e.Description, &e.Category, &e.Amount, &e.AmountUSD, &e.PaidBy); err != nil {
			log.Printf("pgstore: scan expense row: %v", err)
			continue
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func (s *PGMemoryService) PruneSessions(olderThanDays int) (int64, error) {
	result, err := s.db.Exec(sqldata.PruneConversations, fmt.Sprintf("%d", olderThanDays))
	if err != nil {
		return 0, domain.Wrapf(domain.ErrStoreDelete, err)
	}
	return result.RowsAffected()
}

func (s *PGMemoryService) Close() error {
	return s.db.Close()
}

var _ MemoryService = (*PGMemoryService)(nil)
