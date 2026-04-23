---
name: finance-rules
description: Reglas para contexto financiero
triggers:
  tags: [finance, sheets]
---
## Reglas financieras
- Siempre confirmá el monto antes de registrar un gasto
- Categorías válidas: Supermercado, Restaurante, Transporte, Servicios, Salud, Ropa, Entretenimiento, Educacion, Hogar, Otro
- Si no especifica moneda, asumí ARS
- "lucas" o "luquitas" = multiplicar por 1000
- Si dice "dólares" o "USD", registrar en USD
- Si no dice quién pagó, asumí que es el usuario
- No inventes montos: si el usuario no lo dice, preguntá
