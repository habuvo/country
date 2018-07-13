#### Country by IP with public API servers

Get requests on :8081 and check country name by reauesting list of public servers.
Catched results cached in Redis with expire time for repeated requests.

##### Parameters

Parameters set in *configuration.json* file that must present in execute directory.

Fields in configuration file:

ExpireTime  in seconds
  
Providers - array of public API services
    request constructor : PreReqURL + ip + PostReqURL  // depends of API
    KeysInResponce  - array of responce JSON key in hierarhical order
    MaxReqPerMinute - beware of anti-spam banning 




