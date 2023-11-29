## Start the connection
var ws = new WebSocket('ws://localhost:8080/ws');

# setup message

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


## create
ws.send(JSON.stringify({
    operation: "create"
}));

## start
ws.send(JSON.stringify({
    operation: "start",
    containerID: "d2d3888910834aa8d790ee693bc8f9c06a8f8516d032e0bda0892196c455c9ff"
}));

##
ws.send(JSON.stringify({
  operation: "list"
}));

##
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "update"]
}));


##
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "add", "python3"]
}));


##
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "add", "git"]
}));


##
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "add", "curl"]
}));

##
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["curl", "https://example.com"]
}));