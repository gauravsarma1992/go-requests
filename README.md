# Golang Requests

Use this to automate calling HTTP calls by just defining them in a config file

## Config file
```json
{
  "config": {
    "base_url": "http://localhost:9090",
    "stats_folder": "/tmp/stats",
    "auth_mechanism": "",
    "api_key": "",
    "cookies": {
      "Cookie-1": ""
    }
  },
  "requests": [{
    "api": "csv_folder",
    "method": "GET",
    "query_params": [{
      "key": "folder_name",
      "value_type": "static",
      "value": "%2Ftmp%2Fchaos-stats"
    }]
  }]
}
```

The `base_url` is the server against which we want to make the calls.
The `requests` list is the set of APIs against which continuous calls will be made.
The requests are made in parallel and then continued till the channel is closed.
