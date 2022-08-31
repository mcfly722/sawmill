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

var source = FileTails("/var/log/messages").SetQueryIntervalSec(10).FromStart(false)

var influxDB = InfluxDB("https://influx.local.domain:8696").SetAuthByToken(os.Getenv('AUTH_TOKEN')).BatchSize(100).MinIntervalSec(10)

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


  influxDB.PushToBatch(obj)
}

stringByStringLogFlow(src, parser)

```
### available API functions

#### Console
```
console.log(msg string)
```
writes message to output

#### Environment Varaiables
Obtain Environment Variable value.
```
os.Getenv(variableName string)
```
#### Timing
Parse time from string based on specified layout
```
loggedTime = time.Parse(layout string, time.Now())
```
Parse using RFC standard
```
loggedTime = time.Parse(time.RFC3339, time.Now())
```
(all available standards are here: https://pkg.go.dev/time#pkg-constants)
<br>

Compare and Add Milliseconds
```
if loggedTime > time.Now().AddMilliseconds(-60*1000) {
  ...write to log...
}
```
Format time back to string
```
time.Format(layout string)
```
### flattening hierarchies
On pushing object to destination there could be a tree object. If destination system does not support fields hierarchy (like InfluxDB/Elastic), all upper levels would be automatically flattened.<br>
For example:
Object is:
```
{
  "source":{
    "podName" : "ping-test-1e8he",
    "podIP" : "192.168.0.49"
  },
  "destination":{
    "podName" : "ping-test-6845e",
    "podIP" : "192.168.0.50"
  }
}
```
would be automatically flattened to:
```
{
  "source.podName" : "ping-test-1e8he",
  "source.podIP" : "192.168.0.49",
  "destination.podName" : "ping-test-6845e",
  "destination.podIP" : "192.168.0.50"
}
```
If some field have array type, it would be automatically type-casted to string using JSON notation.<br>
For example:
```
{
  "TestArray":[1,2,3,{"some":"object"}]
}
```
would be automatically converted to:
```
{
  "TestArray":"[1,2,3,{\"some\":\"object\"}]"
}
```
