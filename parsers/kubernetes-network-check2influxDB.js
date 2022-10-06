// parser for /var/log/containers/kubernetes-network-check-*.log
//
// sample of string:
// {"log":"{\"Timestamp\":\"2022-10-02T17:54:12Z\",\"Source\":{\"PodName\":\"kubernetes-network-check-8tmtd\",\"PodIP\":\"10.42.16.5\",\"HostName\":\"kub-test-mdcnode1\",\"HostIP\":\"10.64.16.127\"},\"Destination\":{\"PodName\":\"kubernetes-network-check-fs7bc\",\"PodIP\":\"10.42.144.6\",\"HostName\":\"kub-test-mdcnode5\",\"HostIP\":\"10.64.17.141\"},\"Message\":\"64 bytes from 10.42.144.6: icmp_seq=44001 ttl=64 time=0.340 ms\",\"Elapsed_ms\":0.34,\"Success\":true}\n","stream":"stdout","time":"2022-10-02T17:54:12.339059517Z"}

const parseMessagesOnlyForLastNSeconds = 5 * 60; // 5 minutes

var stat = {};

function messagesParser(str){
  try {
    if (str === "") {
      stat.empty++
    } else {

      kubernetes = JSON.parse(str)

      if ((kubernetes.log.indexOf("current pod:")===0) || (kubernetes.log.indexOf("used pods:")===0) || (kubernetes.log.indexOf("pinger for")===0) || (kubernetes.log.indexOf("exec:")===0)) {
        stat.notData++
      } else {

        ping = JSON.parse(kubernetes.log)

        if ((new Date(kubernetes.time)).getTime() < Date.now() - parseMessagesOnlyForLastNSeconds * 1000) {
          stat.outdated++
        } else {
          result = {
            Measurement: "kubernetes-network-check",
            Tags: {
              "ClusterName"         :OS.Getenv("CLUSTER_NAME"),
              "Source.PodName"      :ping.Source.PodName,
              "Source.PodIP"        :ping.Source.PodIP,
              "Source.HostName"     :ping.Source.HostName,
              "Source.HostIP"       :ping.Source.HostIP,
              "Destination.PodName" :ping.Destination.PodName,
              "Destination.PodIP"   :ping.Destination.PodIP,
              "Destination.HostName":ping.Destination.HostName,
              "Destination.HostIP"  :ping.Destination.HostIP,
              "Success"             :ping.Success
            },
            Fields: {
              "Elapsed_ms" :ping.Elapsed_ms
            },
            Timestamp: Date.parse(kubernetes.time)
          }

          stat.parsed++

          return result
        }
      }
    }

  } catch (err) {
    stat.lastException = {
      msg:err.message,
      str:kubernetes.log
    }

    stat.exceptions++
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

var parser = Parser.NewString2JSObject(messagesParser)
  .SetQueueStringsSize(25600)
  .SendTo(influxDB)

var watcher = FilesTails.NewWatcher("kubernetes-network-check-*.log")
  .SetFilesPath("/var/log/containers")
  .SetRelistFilesIntervalMS(10000)
  .SetReadFileIntervalMS(5000)
  .SetQueueStringsSize(25600)
  .SendTo(parser)

function resetStat(){
  stat = {
    parsed: 0,
    exceptions: 0,
    outdated:0,
    empty:0,
    notData:0,
    lastException:{}
  }
}

function showStat(){
  Console.Log(JSON.stringify(stat))
  resetStat()
}

Scheduler.NewTicker(10*1000, showStat).Start()

Console.Log("v1.9")
Console.Log("/var/log/containers/kubernetes-network-check-*.log -> InfluxDB parser started")
