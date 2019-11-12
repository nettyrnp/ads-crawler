package repository

import (
	migrate "github.com/rubenv/sql-migrate"
)

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id: "00001_initial_migration",
			Up: []string{
				"CREATE TYPE acctypes AS ENUM ('direct', 'reseller');",

				`CREATE TABLE portal (
                       id serial primary key not null,
                       raw_url text,
                       canonical_name text unique not null,
                       protocol text not null,
                       email text,
                       phone text,
                       cert_info text,
                       created_at timestamp not null,
                       updated_at timestamp not null,
                       deleted_at timestamp
				);`,
				"CREATE INDEX portal_idx ON portal (canonical_name,protocol);",

				`INSERT INTO portal (protocol,canonical_name,email,phone,cert_info,created_at,updated_at)
					VALUES 
						('http','cnn.com', 'ee1@ee.ee', '+044-1234567', 'files:transport.pem,transport.key;type:pem', NOW(), NOW()),
						('http','gizmodo.com', 'ee2@ee.ee', '+044-1234567', 'files:transport.pem,transport.key;type:pem', NOW(), NOW()),
						('http','nytimes.com', 'ee3@ee.ee', '+044-1234567', 'files:transport.pem,transport.key;type:pem', NOW(), NOW()),
						('https','bloomberg.com', 'ee4@ee.ee', '+044-1234567', 'files:transport.der,transport.key;type:der', NOW(), NOW()),
						('https','wordpress.com', 'ee5@ee.ee', '+044-1234567', 'files:transport.pem,transport.key;type:pem', NOW(), NOW())
					;`,

				`CREATE TABLE provider (
                       id serial primary key not null,
                       domain_name text not null,
                       account_id text not null,
                       account_type acctypes not null,
                       cert_auth_id text,
                       portal_id int,
                       created_at timestamp not null,
                       updated_at timestamp not null,
                       deleted_at timestamp,
					   unique(domain_name,account_id,account_type)
				);`,
				"CREATE INDEX provider_idx ON provider (domain_name,account_id,account_type);",
			},
			Down: []string{
				"DROP INDEX IF EXISTS provider_idx;",
				"DROP TABLE IF EXISTS provider;",
				"DROP INDEX IF EXISTS portal_idx;",
				"DROP TABLE IF EXISTS portal;",
				"DROP TYPE IF EXISTS acctypes;",
			},
		},
	},
}
