# Teleport-Extensions [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)

[Teleport Framework](https://github.com/henrylee2cn/teleport) custom extensions collection.


## Install

```sh
go get -u -f -d github.com/henrylee2cn/tp-ext/...
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
[secure](https://github.com/henrylee2cn/tp-ext/blob/master/plugin-secure)|`import secure "github.com/henrylee2cn/tp-ext/plugin-secure"`|Encrypting/decrypting the packet body

## Protocol

package|import|description
----|------|-----------
[jsonproto](https://github.com/henrylee2cn/tp-ext/blob/master/proto-jsonproto)|`import jsonproto "github.com/henrylee2cn/tp-ext/proto-jsonproto"`|A JSON socket communication protocol
[pbproto](https://github.com/henrylee2cn/tp-ext/blob/master/proto-pbproto)|`import pbproto "github.com/henrylee2cn/tp-ext/proto-pbproto"`|A PTOTOBUF socket communication protocol
[tpV2Proto](https://github.com/henrylee2cn/tp-ext/blob/master/proto-tpV2Proto)|`import tpV2Proto "github.com/henrylee2cn/tp-ext/proto-tpV2Proto"`|Compatible teleport v2 protocol

## Transfer-Filter

package|import|description
----|------|-----------
[md5Hash](https://github.com/henrylee2cn/tp-ext/blob/master/xfer-md5Hash)|`import md5Hash "github.com/henrylee2cn/tp-ext/xfer-md5Hash"`|Provides a integrity check transfer filter

## Module

package|import|description
----|------|-----------
[cliSession](https://github.com/henrylee2cn/tp-ext/blob/master/mod-cliSession)|`import cliSession "github.com/henrylee2cn/tp-ext/mod-cliSession"`|Client session with a high efficient and load balanced connection pool
[websocket](https://github.com/henrylee2cn/tp-ext/blob/master/mod-websocket)|`import websocket "github.com/henrylee2cn/tp-ext/mod-websocket"`|Makes the Teleport framework compatible with websocket protocol as specified in RFC 6455
