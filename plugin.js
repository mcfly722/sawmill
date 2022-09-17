//var ticker = Scheduler.NewTicker(1*1000, function(){
//  Console.Log("tick!")
//}).Start()

function messagesParser(str){
  Console.Log(str)
}

function onNewLogFound(file) {
  Console.Log(str)
}


var influxDB = InfluxDB.NewConnection("https://localhost:8086").Start()

var parser = Parser.NewString2JSObject(messagesParser).SendTo(influxDB)

var watcher = FilesTails.NewWatcher("log")
  .SetFilesPath("../")
  .SetRelistFilesIntervalMS(1000)
  .SetReadFileIntervalMS(1000)
  .SendTo(parser)

Console.Log("started")
