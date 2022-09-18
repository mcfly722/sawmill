function trimQuotes(str){
  return str.replace(/(^"|"$)/g, '').replace(/(^'|'$)/g, '')
}

function listOfParams2Map(params){
  const regexp = /(?<key>[^\s]+)=(?<value>.+?)(?=\s[^\s]+\=|$)/gm
  result = {}
  for(m of (params.matchAll(regexp))){
    result[m[1]]=trimQuotes(m[2])
  }
  return result
}

function messagesParser(str){

  if (!(str === "")) {
    var obj = listOfParams2Map(str)
    return {
      Measurement: "testMeasurement",
      Tags: {
        node:"node1",
        process:"proc"
      },
      Fields: obj,
      Timestamp: obj.time
    }
  }
}

var influxDB = InfluxDB.NewConnection("http://localhost:8086")
  .SetAuthByToken(OS.Getenv("INFLUXDB_TOKEN"))
  .SetOrganization("home")
  .SetBucket("NewBucket")
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
