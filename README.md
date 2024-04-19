# day-guide-telegram-bot

### Development
After clone set the following environment variables:
- TELEGRAM_BOT_TOKEN
- OPEN_AI_TOKEN
- OPEN_WEATHER_MAP_API_KEY
- OPEN_EXCHANGE_RATES_APP_ID

To start the DB:
`docker-compose up -d db`

To start the server:
```
go run main.go
```
Then post a message to the bot.

## Work with PostgreSQL using psql

Switch to the postgres user:
> su postgres

Run psql:
> psql

Display databases:
> \l

Connect to the app database:
> \c app

List tables inside public schemas:
> \dt

Exit psql:
> \q

Logout from the postgres session:
> exit
