// parser for /var/log/containers/kubernetes-network-check-*.log
//
// sample of string:
// {"log":"{\"Timestamp\":\"2022-10-02T17:54:12Z\",\"Source\":{\"PodName\":\"kubernetes-network-check-8tmtd\",\"PodIP\":\"10.42.16.5\",\"HostName\":\"kub-test-mdcnode1\",\"HostIP\":\"10.64.16.127\"},\"Destination\":{\"PodName\":\"kubernetes-network-check-fs7bc\",\"PodIP\":\"10.42.144.6\",\"HostName\":\"kub-test-mdcnode5\",\"HostIP\":\"10.64.17.141\"},\"Message\":\"64 bytes from 10.42.144.6: icmp_seq=44001 ttl=64 time=0.340 ms\",\"Elapsed_ms\":0.34,\"Success\":true}\n","stream":"stdout","time":"2022-10-02T17:54:12.339059517Z"}


const parseMessagesOnlyForLastNSeconds = 5 * 60; // 5 minutes

function messagesParser(str){
  try {
    if (!(str === "")) {

      kubernetes = JSON.parse(str)
      ping = JSON.parse(kubernetes.log)

      if ((new Date(kubernetes.time)).getTime() > Date.now() - parseMessagesOnlyForLastNSeconds * 1000) {
        result = {
          Measurement: "kubernetes-network-check",
          Tags: {
            "Source.PodName"      :ping.Source.PodName,
            "Source.PodIP"        :ping.Source.PodIP,
            "Source.HostName"     :ping.Source.HostName,
            "Source.HostIP"       :ping.Source.HostIP,
            "Destination.PodName" :ping.Destination.PodName,
            "Destination.PodIP"   :ping.Destination.PodIP,
            "Destination.HostName":ping.Destination.HostName,
            "Destination.HostIP"  :ping.Destination.HostIP,
          },
          Fields: {
            "Elapsed_ms" :ping.Elapsed_ms,
            "Success"    :ping.Success
          },
          Timestamp: Date.parse(kubernetes.time)
        }

        return result
      }
    }

  } catch {}
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

var watcher = FilesTails.NewWatcher("kubernetes-network-check-*.log")
  .SetFilesPath("/var/log/containers")
  .SetRelistFilesIntervalMS(10000)
  .SetReadFileIntervalMS(5000)
  .SendTo(parser)

Console.Log("/var/log/containers/kubernetes-network-check-*.log -> InfluxDB parser started")
