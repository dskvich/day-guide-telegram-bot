# day-guide-telegram-bot

### Development
After clone set the `TELEGRAM_BOT_TOKEN` and `GPT_TOKEN` environment variables.

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
