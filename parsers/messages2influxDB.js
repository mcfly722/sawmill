// parser for /var/log/messages Centos 7
//
// sample:
// Aug 28 03:31:36 kub-test-node1 dockerd: time="2022-08-28T03:31:36.810790504+03:00" level=info msg="ignoring event" moudle=libcontainerd namespace=moby topic=/tasks/delete type="*events.TaskDelete"


function getTags(params){
  const regexp = /(?<key>[^\s:]+)/gm
  result = []
  for(m of (params.matchAll(regexp))){
    result.push(m[0])
  }
  return result
}

function trimQuotes(str){
  return str.replace(/(^"|"$)/g, '').replace(/(^'|'$)/g, '')
}

function getFields(params){
  const regexp = /(?<key>[^\s]+)=(?<value>.+?)(?=\s[^\s]+\=|$)/gm
  result = {}
  for(m of (params.matchAll(regexp))){
    result[m[1]]=trimQuotes(m[2])
  }
  return result
}

const parseMessagesOnlyForLastNSeconds = 5 * 60; // 5 minutes

function messagesParser(str){

  if (!(str === "")) {

    var tags = getTags(str)
    var fields = getFields(str)

    var activeTags = {
      "node"        :tags[5],
      "process"     :tags[6]
    }

    for (key in fields) {
      activeTags[key] = fields[key]
    }

    if ((new Date(fields.time)).getTime() > Date.now() - parseMessagesOnlyForLastNSeconds * 1000) {

      result = {
        Measurement: "/var/log/messages",
        Tags: activeTags,
        Fields: {"zero":0},
        Timestamp: Date.parse(fields.time)
      }

      return result
    }
  }
  
}

var influxDB = InfluxDB.NewConnection(OS.Getenv("INFLUXDB_URL"))
  .SetAuthByToken(OS.Getenv("INFLUXDB_TOKEN"))
  .SetOrganization(OS.Getenv("INFLUXDB_ORGANIZATION"))
  .SetBucket(OS.Getenv("INFLUXDB_BUCKET"))
  .SetSendMaxBatchSize(3000)
  .SetSendTimeoutMS(2000)
  .SetSendIntervalMS(5000)
  .OnSendError(function(errorMsg, batch){
    Console.Log("error:"+errorMsg+" (batchSize="+batch.length+")")
  })
  .OnSendSuccess(function(batch){
    Console.Log("success send (batchSize="+batch.length+")")
  })
  .Start()

var parser = Parser.NewString2JSObject(messagesParser).SendTo(influxDB)

var watcher = FilesTails.NewWatcher("messages")
  .SetFilesPath("/var/log")
  .SetRelistFilesIntervalMS(1000)
  .SetReadFileIntervalMS(1000)
  .SendTo(parser)

Console.Log("/var/log/messages -> InfluxDB parser started")
