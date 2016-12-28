gopli
========
Database backup between remote hosts (or local) written in Golang.

## Feature

- High-speed parallel data fetching with goroutine concurrency
- Reuse options of connection with TOML configuration
- Gopli will release you from an annoying replication setting

# TODO
- [ ] Currently MySQL only. so adopt to other management systems
- [ ] Data mask for password, credit-card number, etc...
- [ ] Response packet regulation and compression for fetched data

## Install
```
go get github.com/timakin/gopli
```

## Usage
Write down setting file in toml.
```
[database]
  [database.local]
  host = "localhost"
  management_system = "mysql"
  name = "app_development"
  user = "root"
  password = ""

  [database.staging]
  host = "xxx.xxx.xxx.xxx"
  management_system = "mysql"
  name = "app_staging"
  user = "root"
  password = ""

  [database.production]
  host = "yyy.yyy.yyy.yyy"
  management_system = "mysql"
  name = "app_production"
  user = "root"
  password = ""

[ssh]
  [ssh.local]
  host = "localhost" # or "127.0.0.1"

  [ssh.staging]
  host = "xxx.xxx.xxx.xxx"
  port = "22"
  user = "timakin"
  key = "~/.ssh/id_rsa_staging"

  [ssh.production]
  host = "yyy.yyy.yyy.yyy"
  port = "22"
  user = "remoteuser"
  key = "~/.ssh/id_rsa_prod"

```

```
gopli sync -from production -to staging -c config/gopli.toml
```
