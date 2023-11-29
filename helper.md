## Start the connection
```javascript
var ws = new WebSocket('ws://localhost:8080/ws');
```

# setup message
```javascript
ws.onmessage = function(event) {
    var data = JSON.parse(event.data);
    if (Array.isArray(data)) {
        data.forEach(function(container) {
            console.log('Container ID: ' + container.ID);
            console.log('Image: ' + container.Image);
            console.log('Command: ' + container.Command);
            console.log('Created: ' + container.Created);
            console.log('Status: ' + container.Status);
            console.log('Names: ' + container.Names.join(', '));
            console.log('---');
        });
    } else {
        console.log('Server: ' + event.data);
    }
};
```

## create
```javascript
ws.send(JSON.stringify({
    operation: "create"
}));
```


## start
```javascript
ws.send(JSON.stringify({
    operation: "start",
    containerID: "your-container-id"
}));
```


##
```javascript
ws.send(JSON.stringify({
  operation: "list"
}));
```

##
```javascript
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "update"]
}));
```

##
```javascript
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "add", "python3"]
}));
```

##
```javascript
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "add", "git"]
}));
```


##
```javascript
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "add", "curl"]
}));
```


##
```javascript
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["curl", "https://example.com"]
}));
```

