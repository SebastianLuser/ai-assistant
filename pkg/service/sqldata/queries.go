package sqldata

import _ "embed"

//go:embed postgres/insert/save_memory.sql
var SaveMemory string

//go:embed postgres/insert/save_conversation.sql
var SaveConversation string

//go:embed postgres/insert/log_habit.sql
var LogHabit string

//go:embed postgres/select/all_memories.sql
var SelectMemories string

//go:embed postgres/select/search_fts.sql
var SearchFTS string

//go:embed postgres/select/load_conversation.sql
var LoadConversation string

//go:embed postgres/select/count_habits.sql
var CountHabit string

//go:embed postgres/select/habit_dates.sql
var HabitDates string

//go:embed postgres/select/habits_today.sql
var HabitsToday string

//go:embed postgres/select/list_expenses.sql
var ListExpenses string

//go:embed postgres/delete/delete_memory.sql
var DeleteMemory string

//go:embed postgres/delete/clear_conversation.sql
var ClearConversation string

//go:embed postgres/delete/prune_conversations.sql
var PruneConversations string

//go:embed postgres/insert/upsert_catalog.sql
var UpsertCatalog string

//go:embed postgres/select/catalog_list.sql
var CatalogList string

//go:embed postgres/select/catalog_get.sql
var CatalogGet string
