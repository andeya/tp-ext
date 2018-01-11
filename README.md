# Teleport-Extensions [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)

[Teleport Framework](https://github.com/henrylee2cn/teleport) custom extensions collection.


## Install

```sh
go get -u github.com/henrylee2cn/tp-ext/...
```

## Codec

package|import|description
----|------|-----------

## Plugin

package|import|description
----|------|-----------
[binder](https://github.com/henrylee2cn/tp-ext/blob/master/plugin-binder)|`import binder "github.com/henrylee2cn/tp-ext/plugin-binder"`|Parameter Binding Verification for Struct Handler
[heartbeat](https://github.com/henrylee2cn/tp-ext/blob/master/plugin-heartbeat)|`import heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"`|A generic timing heartbeat plugin
[ignoreCase](https://github.com/henrylee2cn/tp-ext/blob/master/plugin-ignoreCase)|`import ignoreCase "github.com/henrylee2cn/tp-ext/plugin-ignoreCase"`|Dynamically ignoring the case of path

## Protocol

package|import|description
----|------|-----------

## Transfer-Filter

package|import|description
----|------|-----------

## Sundry

package|import|description
----|------|-----------
[cliSession](https://github.com/henrylee2cn/tp-ext/blob/master/sundry-cliSession)|`import cliSession "github.com/henrylee2cn/tp-ext/sundry-cliSession"`|Client session which has connection pool
[websocket](https://github.com/henrylee2cn/tp-ext/blob/master/sundry-websocket)|`import websocket "github.com/henrylee2cn/tp-ext/sundry-websocket"`|Makes the Teleport framework compatible with websocket protocol as specified in RFC 6455
