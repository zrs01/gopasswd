# gopasswd - Password change tool in batch

## Version 2.0.0
> 2022-12-06
- Support multiple users for a host
- Structure of data.json is changed

## Version 1.2.0
> 2020-09-17
- Upgraded package "github.com/urfave/cli" to "github.com/urfave/cli/v2"
- Replaced cli password with interactive input

## Version 1.1.0
> 2019-11-11
- encrypt the password when save to the file
- add command line option to check the host conectivity

## Testing Environment Setup

```sh
# Start centos
docker run -d --name centos-ssh -p 2222:22 --log-driver json-file --log-opt max-size=1m jdeathe/centos-ssh:centos-7

# Get into container
docker exec -it centos-ssh bash

# Reinstall crack dictionary
$ yum reinstall -y cracklib-dicts

# Modify /etc/sshd_config
$ vi /etc/sshd_config
# change
PasswordAuthentication no -> yes

# Create a user for testing
$ useradd user
$ passwd user

# restart docker
docker restart centos-ssh

# verify user login
$ ssh -p 2222 user@127.0.0.1
```