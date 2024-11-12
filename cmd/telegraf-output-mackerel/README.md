# telegraf-output-mackerel

Telegraf output plugin for Mackerel.io

## How to Use

1. Append the following configuration to `telegraf.conf`.

```toml
[[outputs.exec]]
  command = ["/path/to/telegraf-output-mackerel", "--config", "/path/to/telegraf-output-mackerel.conf"]
  timeout = "10s"
```

2. Create a config file that `telegraf-output-mackerel` will refer to. This is confusing, but must be in a separate file from `telegraf.conf`.

Here we create a file named `/path/to/telegraf-output-mackerel.conf`. A sample of this file can be found [here](https://github.com/SlashNephy/telegraf-output-mackerel/blob/master/plugins/outputs/mackerel/sample.conf).

```toml
[[outputs.mackerel]]
  # Required
  # API keys can be issued on the Mackerel dashboard.
  # Alternatively, you can set it via the $MACKEREL_API_KEY environment variable.
  api_key = ""

  # Either of the following is required
  # Specify host_id if you want the metrics to be associated with a host,
  # or service_name if you want them to be associated with a service.
  # Alternatively, you can set it via the environment variable $MACKEREL_HOST_ID or $MACKEREL_SERVICE_NAME.
  host_id = ""
  service_name = ""

  # Optional
  # Specify a prefix for the metric name.
  # A unique prefix that does not conflict with other metric names is recommended.
  # Alternatively, you can set it via the $MACKEREL_METRIC_PREFIX environment variable.
  metric_prefix = "telegram"
```

3. Restart telegraf instance.
