# YAMS

**Y**et **A**nother **M**ocking **S**erver written in [Go](https://golang.org/) and [Vue 2](https://vuejs.org), that can
be configured with scripts in [Lua 5.1](https://www.lua.org/manual/5.1/). YAMS can act as a basic backend, which
supports request analysis, output buffering, simple database via storage variables, sessions, and debugging.

## Install

1. Make sure you have Go 1.9+ and Node 6+ installed in the system.
2. Check `GOPATH` is configured to `/home/user/go` to make things easy.
3. To download and deploy YAMS code run the following command:

       $ go get -u github.com/lokhman/yams

4. Then build frontend using `npm` with the following:

       $ cd $GOPATH/src/github.com/lokhman/yams/public/console
       $ npm run dist

5. YAMS requires PostgreSQL 10+ database. SQL schema can be found [here](docs/yams.sql).
6. Configure and start [Supervisor](http://supervisord.org) using the command below. Supervisor is required to keep YAMS
running in the background and to secure server restart. Example configuration can be found
[here](docs/supervisor.conf).

       $ sudo supervisorctl start yams

7. Finally, install and configure _nginx_ or _Apache_ web server to act as a proxy and load balancer. Example
configuration for _nginx_ can be found [here](docs/nginx.conf).

YAMS exposes two ports: _8086_ to proxy the requests and _8087_ for administration console (ports can be changed with
`--proxy-addr` and `--console-addr` CLI flags).

## Upgrade

1. Stop Supervisor process and make sure `yams` is unloaded from the memory.

       $ sudo supervisorctl stop yams

2. Run `go get` command to download the newest version of YAMS code:

       $ go get -u github.com/lokhman/yams

3. If required rebuild frontend with `npm`:

       $ cd $GOPATH/src/github.com/lokhman/yams/public/console
       $ npm run dist

4. Start back Supervisor process:

       $ sudo supervisorctl start yams

## License

YAMS is available under the MIT license. The included LICENSE file describes this in detail.
