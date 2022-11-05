# APIs

## ⏰ Job APIs

### Create a job
`POST /job/:db/:collection`
```jsonc
Request:
{
    "id": "nxz123bnj",
    "trigger_time": 1667659342626,
    "meta": {
        // Any json that you want to pass on to the reciepent
    },
    "route": "gameServer" // The reciepient route
}

Response 200: 
{
    "status": "ok"
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```

### Fetch a job
`GET /job/:db/:collection/:id`
```jsonc
Response 200:
{
    "id": "nxz123bnj",
    "trigger_time": 1667659342626,
    "meta": {
        // Any json that you want to pass on to the reciepent
    },
    "route": "gameServer" // The reciepient route
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```

### Update a job
`PUT /job/:db/:collection/:id`
```jsonc
Request:
{
    "id": "nxz123bnj",
    "trigger_time": 1667659342626,
    "meta": {
        // Any json that you want to pass on to the reciepent
    },
    "route": "gameServer" // The reciepient route
}

Response 200: 
{
    "status": "ok"
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```

### Delete a job
`DELETE /job/:db/:collection/:id`
```jsonc
Response 200: 
{
    "status": "ok"
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```

## 🛺 Route APIs

### Create a route
`POST /route/:db`
```jsonc
Request:
{
    "id": "gameServer",
    "type": "REST",
    "webhook_url": "gameserver-dev-1.myorg.com/timer?action=endgame" // Your URL webhook
}

Response 200: 
{
    "status": "ok"
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```

### Fetch a route
`GET /route/:db/:id`
```jsonc
Response 200:
{
    "id": "gameServer",
    "type": "REST",
    "webhook_url": "gameserver-dev-1.myorg.com/timer?action=endgame" // Your URL webhook
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```

### Update a route
`PUT  /route/:db/:id`
```jsonc
Request:
{
    "id": "gameServer",
    "type": "REST",
    "webhook_url": "gameserver-dev-1.myorg.com/timer?action=endgame" // Your URL webhook
}

Response 200: 
{
    "status": "ok"
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```

### Delete a route
`DELETE /route/:db/:id`
```jsonc
Response 200: 
{
    "status": "ok"
}

Response 400:
{
    "error": "Human readable reason", // For humans
    "code": "E001" // For robots
}
```