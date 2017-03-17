(function(){
    var ws = new WebSocket(`ws://${location.host}/ws`);

    ws.onmassage = function(e){
        // instance name
        // status
        // nat ip
        // network ip

        console.log(JSON.parse(e.data))
    };



})();
