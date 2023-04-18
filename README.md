# start-service

- every start
- quasi Meldeergebnis
- results from Livetiming/Protocol

## Models

### Start

- id


## API Endpoints

```
[GIN-debug] GET    /start                    --> sr-start/start-service/controller.getStarts (3 handlers)
[GIN-debug] GET    /start/:id                --> sr-start/start-service/controller.getStart (3 handlers)
[GIN-debug] DELETE /start/:id                --> sr-start/start-service/controller.removeStart (3 handlers)
[GIN-debug] POST   /start                    --> sr-start/start-service/controller.addStart (3 handlers)
[GIN-debug] PUT    /start                    --> sr-start/start-service/controller.updateStart (3 handlers)
[GIN-debug] GET    /actuator                 --> sr-start/start-service/controller.actuator (3 handlers)
```