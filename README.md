# Tableau-Exporter-Prometheus
Export Tableau metrics for Prometheus

## Build the docker image
```
docker build -t tableau_exporter .
```
## Run the docker image
```
docker run --rm -p 9030:9030 -e DATABASE_HOST=tableauServer -e DATABASE_PASSWORD=tableauPostgreSQLUser -e DATABASE_PASSWORD=tableauPostgreSQLPassword tableau_exporter:latest
```

## Configuration
```yaml
database:
  host: "localhost"
  port: 8060
  name: "workgroup"
  user: "readonly"
  password: "password"

scrapeIntervalSeconds: 5
port: 9030
```

