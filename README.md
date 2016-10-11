Cylinder
========
Database backup between remote hosts (or local) written in Golang.

## Install
```
go get -u github.com/timakin/cylinder
```

## Usage
Write down setting file in toml.
```
[database]
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
cylinder sync -from production -to staging -c config/cylinder.toml
```
