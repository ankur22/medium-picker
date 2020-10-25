# medium-picker

![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/b06d2ab8a21941b78abc54eafd1941e4)](https://app.codacy.com/gh/ankur22/medium-picker?utm_source=github.com&utm_medium=referral&utm_content=ankur22/medium-picker&utm_campaign=Badge_Grade)
![Lint everything](https://github.com/ankur22/medium-picker/workflows/Lint%20everything/badge.svg)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B20992%2Fgithub.com%2Fankur22%2Fmedium-picker.svg?type=shield)](https://app.fossa.com/projects/custom%2B20992%2Fgithub.com%2Fankur22%2Fmedium-picker?ref=badge_shield)
[![Build Status](https://travis-ci.com/ankur22/medium-picker.svg?branch=main)](https://travis-ci.com/ankur22/medium-picker)
[![codecov](https://codecov.io/gh/ankur22/medium-picker/branch/main/graph/badge.svg?token=T5NKEL12CW)](https://codecov.io/gh/ankur22/medium-picker)

Pick a blog/news site to catch up on so you don't feel like you need to read everything on the internet

## How to build it

```shell
CGO_ENABLED=0 go build -ldflags="-X 'main.Version=`cat VERSION.txt`' -X 'main.Commit=`git rev-parse HEAD`'" -o ./app cmd/server/main.go
```

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

| Method | Endpoint                        | Query | Request Body         | Reponse Body                           | Success Code | Failures | Description               |
|--------|---------------------------------|-------|----------------------|----------------------------------------|--------------|----------|---------------------------|
| POST   | /v1/user                        | -     | {"username": string} | {"userId": "string"}                   | 201          | 400 409  | Create account            |
| PUT    | /v1/user/login                  | -     | {"username": string} | {"userId": "string"}                   | 200          | 400 404  | Login                     |
| POST   | /v1/user/{userID}/medium        | -     | {"source": string}   | -                                      | 204          | 400 409  | Add a new medium source   |
| GET    | /v1/user/{userID}/medium        | p=int | -                    | [{"source": string, "Id": string, "nextPage": int}]   | 200      | 400      | Get all the sources (paginated) |
| DELETE | /v1/user/{userID}/medium/{Id}   | -     | -                    | -                                      | 204          | 404      | Delete a medium source    |
| GET    | /v1/user/{userID}/medium/choose | n=int | -                    | [{"url": "string"}]                    | 200          | 400 404  | Get n medium urls to read |

## Store Schema

### Medium Sources

| Name         | Type   | Description                               |
|--------------|--------|-------------------------------------------|
| Source       | string | The URL to the site. It's the primary key |
| Id           | string | A UUID                                    |
| Hash         | string | The hash of the webpage                   |
| Multiplier   | float  | Increase the chance of it being picked    |
| CreatedDate  | date   | When the record was created               |
| ModifiedDate | date   | When the record was modified              |
| Hit          | int    | Number of times this record was picked    |
| UserId       | string | The user token this is associated with    |

### Users

| Name         | Type   | Description                  |
|--------------|--------|------------------------------|
| Email        | string | The user's email address     |
| UserId       | string | A UUID. It's the primary key |
| CreatedDate  | date   | When the record was created  |
| ModifiedDate | date   | When the record was updated  |

## License

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B20992%2Fgithub.com%2Fankur22%2Fmedium-picker.svg?type=large)](https://app.fossa.com/projects/custom%2B20992%2Fgithub.com%2Fankur22%2Fmedium-picker?ref=badge_large)
