// parser for /var/log/messages Centos 7
//
// sample of string:
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

function messagesParser(str){

  if (!(str === "")) {

    var tags = getTags(str)
    var fields = getFields(str)

    result = {
      Measurement: "testMeasurement",
      Tags: {
        node:tags[5],
        process:tags[6]
      },
      Fields: fields,
      Timestamp: Date.parse(fields.time)
    }

    return result
  }
}

var influxDB = InfluxDB.NewConnection("http://localhost:8086")
  .SetAuthByToken(OS.Getenv("INFLUXDB_TOKEN"))
  .SetOrganization("home")
  .SetBucket("NewBucket3")
  .SetSendMaxBatchSize(3)
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

var watcher = FilesTails.NewWatcher("log")
  .SetFilesPath("../")
  .SetRelistFilesIntervalMS(1000)
  .SetReadFileIntervalMS(1000)
  .SendTo(parser)

Console.Log("started")
