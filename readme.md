# sawmill
![Status: techdoc](https://img.shields.io/badge/status-%20techdoc-yellow.svg)

Module reads log files, tries to parse it with specified JavaScript custom parser and sends it to warehouse (InfluxDB/ElasticSearch).<br><br>
Key advantage is to use JavaScript language for parsing.<br><br>
Parsing function obtain line by line, combines them together with required logic and outputs complete objects to warehouse.
User is not limited with Grok or RegExps notations. Instead he uses full JavaScript language power.<br> 
