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

var dst = InfluxDB("https://influx.local.domain:8696").SetAuthByToken(GetEnv('AUTH_TOKEN')).BatchSize(100).MinIntervalSec(10)

function listOfParams2Map(params){
  const regexp = /(?<key>[^\s]+)=(?<value>.+?)(?=\s[^\s]+\=|$)/gm
  result = {}  
  for(const m of (params.matchAll(regexp))){
    result[m[1]]=m[2].replace(/(^"|"$)/g, '')
  }
  return result
}

# str = 'Aug 28 03:31:36 kub-test-node1 dockerd: time="2022-08-28T03:31:36.810790504+03:00" level=info msg="ignoring event" moudle=libcontainerd namespace=moby topic=/tasks/delete type="*events.TaskDelete"'
function parser(str){



}

newLogFlow(src,parser,dst)
```
### flattering json
