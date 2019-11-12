Ads Crawler
========================
Backend part of the Ads Crawler.
Ads Crawler is a service that collects and stores the infos about ads providers on the given portals (e.g. 'www.wordpress.com'). It also provides an API for quering this info.

## Installation

#### Install Go
https://golang.org/doc/install

#### Download the App repository
```
go get github.com/nettyrnp/ads-crawler
```

#### Install and Start Postgres
[Ubuntu]
https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-18-04

[MacOS]
brew postgresql-upgrade-database
brew services start postgresql

Create database in postgres
```
createdb crawler_be
```

(optional) Try to connect to the database:
```
psql crawler_be
```

Create tables in the database:
```
make migrate
```

## Running the application
#### Enabling HTTPS mode
In '.env' file change this line to
PROTOCOL=http   // for HTTP mode
or
PROTOCOL=https   // for HTTPS mode

#### Running in Docker:
Build
```
docker build -t ads-crawler -f Dockerfile2 .
```
Run
```
docker run -d -p 8080:8080 ads-crawler
```
Stop
```
docker ps

docker stop <CONTAINER ID>
```

#### Running in Terminal:
```
make run
```
Now visit http://localhost:8080/api/v0/crawler/admin/version and see the App version in your browser. 
(Or https://...  -- if you are running the crawler in HTTPS mode.)


## REST API:
========================
Examples of Postman requests can be found in testdata/nettyrnp-crawler.postman_collection.json

#### Main routes:
    GET localhost:8080/api/v0/crawler/admin/version // to get the crawler API version
    GET localhost:8080/api/v0/crawler/admin/logs    // to get latest part of logs
    GET localhost:8080/api/v0/crawler/portals       // to get the list of portals (e.g. 'www.wordpress.com')
    POST localhost:8080/api/v0/crawler/portals      // to get the list of portals in a filtered, sorted and paged form
    POST localhost:8080/api/v0/crawler/start_poll   // to start gathering of ads providers at the portals
    GET localhost:8080/api/v0/crawler/providers/portal/wordpress.com    // to get the list of ads providers for specific portal


#### Sample CURL request:
in HTTP mode:
```
curl -X POST   http://localhost:8080/api/v0/crawler/start_poll   -H 'cache-control: no-cache'
```
in HTTPS mode:
```
curl -X POST   --cert './localhost+1.pem'   --cert-type PEM   --key './localhost+1-key.pem'   https://localhost:8080/api/v0/crawler/start_poll   -H 'cache-control: no-cache'
```
The GET requests can be executed also in browser.

 
