//var ticker = Scheduler.NewTicker(1*1000, function(){
//  Console.Log("tick!")
//}).Start()

function parser(str){
  Console.Log(str)
}

function onNewLogFound(file) {
  Console.Log(str)
}

var watcher = FilesTails.NewWatcher("log")
  .SetFilesPath("../")
  .SetRelistFilesIntervalMS(1000)
  .SetReadFileIntervalMS(1000)
  .StartWithParser(parser)

Console.Log("started")
