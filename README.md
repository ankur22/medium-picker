# medium-picker

Pick a blog/news site to catch up on so you don't feel like you need to read everything on the internet

## How it works

### Server

1. On schedule get all the sites concurrently
2. Log the failures
3. Log the success and hash the body of the site
4. Update the hashes of the sites

### Client request

1. Order by Hit desc
2. Find the n number of records that were recently changed (based on modified date)
3. Display the n records that are chosen
4. Update the n records Hit count

## REST API

| Method | Endpoint                      | Query                         | Request Body | Reponse Body        | Success Code | Failures |
|--------|-------------------------------|-------------------------------|--------------|---------------------|--------------|----------|
| GET    | /v1/user/{userID}/blog/choose | n=int number of sites to pick | -            | [{"url": "string"}] | 200          | 404      |

## Store Schema

### Blogs

| Name         | Type   | Description                               |
|--------------|--------|-------------------------------------------|
| Source       | string | The URL to the site. It's the primary key |
| Hash         | string | The hash of the webpage                   |
| Multiplier   | float  | Increase the chance of it being picked    |
| CreatedDate  | date   | When the record was created               |
| ModifiedDate | date   | When the record was modified              |
| Hit          | int    | Number of times this record was picked    |
| UserId       | string | The user token this is associated with    |

### User

| Name         | Type   | Description                  |
|--------------|--------|------------------------------|
| Email        | string | The user's email address     |
| UserId       | string | A UUID. It's the primary key |
| CreatedDate  | date   | When the record was created  |
| ModifiedDate | date   | When the record was updated  |