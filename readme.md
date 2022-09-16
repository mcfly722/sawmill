# sawmill
![Status: in progress](https://img.shields.io/badge/status-in%20progress-yellow.svg)

Module reads log files, tries to parse it with specified JavaScript custom parser and sends it to warehouse (InfluxDB/ElasticSearch).<br><br>
Key advantage is to use JavaScript language for parsing.<br><br>
Parsing function obtain line by line, combines them together with required logic and outputs complete objects to warehouse.
User is not limited with Grok or RegExps notations. Instead, he uses full JavaScript language power.<br>

### how to use
```
.\sawmill -config parser.js
```
in <b>parser.js</b> you have to define all parsing logic and outputs.
### creating your own parser.js
```
function trimQuotes(str){
  return str.replace(/(^"|"$)/g, '').replace(/(^'|'$)/g, '')
}

function listOfParams2Map(params){
  const regexp = /(?<key>[^\s]+)=(?<value>.+?)(?=\s[^\s]+\=|$)/gm
  result = {}  
  for(m of (params.matchAll(regexp))){
    result[m.groups.key]=trimQuotes(m.groups.value)
  }
  return result
}

# example
# str = 'Aug 28 03:31:36 kub-test-node1 dockerd: time="2022-08-28T03:31:36.810790504+03:00" level=info msg="ignoring event" moudle=libcontainerd namespace=moby topic=/tasks/delete type="*events.TaskDelete"'

var influxDB = InfluxDB("https://influx.local.domain:8696")
  .SetAuthByToken(os.Getenv('AUTH_TOKEN'))
  .BatchSize(100)
  .BatchesQueueSize(100)
  .DelaysBetweenSendingMS(1000)
  .Start()

function parser(str){
    obj = listOfParams2Map(str)
    influxDB.Push(obj)
}

TailOfFile("/var/log/messages")
  .SetQueryIntervalMS(10000)
  .FromStart(false)
  .StartWithParser(parser)

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
loggedTime = time.Parse("01/02 03:04:05PM '06 -0700", time.Now())
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
time.Format("01/02 03:04:05PM '06 -0700")
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
