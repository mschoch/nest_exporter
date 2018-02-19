# nest_exporter

A tool to let you integrate your Nest devices with a prometheus system.

## Usage

1. Create a [Nest Developer Account](developers.nest.com)
2. Create a Product
3. Visit the Authorization URL, authorize and note the auth code
4.  Use the `nestauth` tool to turn this into an auth token
5.  Run nest_exporter with this auth token

```
./nest_exporter --token <auth token>
```

Point prometheus to scrape this endpoint (defaults to port 9264)
