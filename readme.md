# sawmill
![Status: techdoc](https://img.shields.io/badge/status-%20techdoc-yellow.svg)

Module reads log files, tries to parse it with specified JavaScript custom parser and sends it to warehouse (InfluxDB/ElasticSearch).<br><br>
Key advantage is to use JavaScript language for parsing.<br><br>
Parsing function obtain line by line, combines them together with required logic and outputs complete objects to warehouse.
User is not limited with Grok or RegExps notations. Instead he uses full JavaScript language power.<br> 

### how to use
```
.\sawmill -config parser.js
```
in <b>parser.js</b> you have to define all parsing logic and outputs.
### creating your own parser.js
```

var src = FileTails("/var/log/messages").SetQueryIntervalSec(10).FromStart(false)

var dst = InfluxDB("https://influx.local.domain:8696").SetAuthByToken(getEnv('AUTH_TOKEN')).BatchSize(100).MinIntervalSec(10)

newLogFlow(src,parser,dst)

)

```
