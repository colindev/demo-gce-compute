(function(factory){

    factory(self.jQuery, self.Metadata, self.Paging);

})(function($, Metadata, Paging){

// jquery plug
$.fn.active = function(){
    this.addClass('active');
    return this;
};
$.fn.unactive = function(){
    this.removeClass('active');
    return this;
};
$.fn.lock = function(){
    if (this.prop('tagName').match(/^(input|button)$/i)) {
        this.attr('readonly', 'readonly');
        this.attr('disabled', 'disabled');
    }
};
$.fn.unlock = function(){
    if (this.prop('tagName').match(/^(input|button)$/i)) {
        this.removeAttr('readonly');
        this.removeAttr('disabled');
    }
};
$.fn.radioButtonBox = function(prefix, conf, onchange){
    var box = this,
        propName = prefix.replace(/[^a-z]*$/, ''),
        rePrefix = new RegExp(`^${prefix}`),
        collect = [], 
        clickHandler = function(e){
        
            var me = this;
            $(collect).each(function(i, btn){
                if (btn === me) {
                    $(this).active();
                    conf.set(propName, me.id.replace(rePrefix, '')).store();
                    return
                }
                $(this).unactive();
            });
    
            if (onchange && onchange.call) {
                onchange.call(me, conf.get(propName));
            }
        };


    $(box).find(`button[id^=${prefix}]`).each(function(){
        collect.push(this);
    }).on('click', clickHandler);

    this.change = function(){
        clickHandler.call(box.find("#"+prefix+conf.get(propName)).get(0) || box.find(`[id^=${prefix}]`).first().get(0));
    };

    return this;
};

var ws = null;
function connWS(url, handler) {
    ws = new WebSocket(url);
    ws.onmessage = handler;
    ws.onopen = function(){console.log('connected')};
    ws.onclose = function(){
        setTimeout(function(){console.log('try reconnect');connWS(url, handler);}, 3000);
    };
}

var conf = new Metadata('config', {
        name: "",
        image: "centos-7-v20170227",
        cpu: "1",
        memory: "1024",
        startup_script: "#!/usr/bin/env bash\n\nyum update -y"
    }), 
    page = (new Paging([
        {path:"/", name: "虛擬機規格"}, 
        // {path:"/machine_type.html", name: "虛擬機規格"}, 
        {path:"/startup_script.html", name:"啟動腳本"},
        {path:"/create.html", name:"虛擬機狀態"}])).bind('button.paging-prev', 'button.paging-next');

$('#catalog').html(conf.get('image')+`<span>${page.current('name')}</span>`);
$(document).ajaxError(function(e, xhr, sets, err){
    alert(xhr.responseText)
});

page.on(['/', '/index.html'], function(){

    var $cpu = $('#cpu'),
        $memory = $('#memory').on('change', function(){
            conf.set('memory', this.value).store();
            $('#display-memory').val(this.value);
        });

    $([1,2,4,6,8,10,12,14]).each(function(i, cpu){
        $cpu.append(`
            <div class="col-xs-4 col-md-3">
                <button id="cpu-${cpu}">CPU ${cpu}</button>
            </div>
        `)
    });

    $('#display-memory').on('change', function(){
        $memory.val(this.value).change();
    })

    $cpu.radioButtonBox('cpu-', conf, function(cpu){
        cpu = parseInt(cpu, 10);

        var min = Math.ceil((cpu * 0.9 * 1024) / 256) * 256 ,
            max = Math.floor((cpu * 6.5 * 1024) / 256) * 256,
            current = parseFloat(conf.get("memory"));

        if (current < min) current = min;

        $memory.attr('min', min).attr('max', max).attr('step', 256).val(current).change();

    }).change();

}).on('/startup_script.html', function(){

    $('#startup-script').val(conf.get("startup_script")).on('change', function(e){
        conf.set("startup_script", this.value).store();
    });
    
    $('input#compute-name').on('change', function(){
        conf.set('name', this.value).store();
    }).val(conf.get('name'));

}).on('/create.html', function(){

var processBox = document.getElementById('process-status');

function onmessage(e){

    var status = JSON.parse(e.data);
    console.log(status)

    if (`compute#instance#${conf.get('name')}` != status.active) {
        return
    }

    processBox.innerText += '\n'+e.data;
    processBox.scrollTop = processBox.scrollHeight;
    if (status.items["status"]) {
        $('#detail-status').text(status.items["status"]);
    }
    if (status.items['network-ip']) {
        $('#detail-network-ip').text(status.items['network-ip']);
    }
    if (status.items['nat-ip']) {
        $('#detail-nat-ip').text(status.items['nat-ip']);
    }

};

connWS(`ws://${location.host}/ws`, onmessage);

$('#detail-cpu').text(conf.get('cpu'));
$('#detail-memory').text(conf.get('memory'));
$('#detail-name').text(conf.get('name'));
$('#btn-create').on('click', function(){

    var $me = $(this),
        data = JSON.parse(localStorage.getItem('config'));

    if (!confirm('確認新增?')) {
        return
    }

    $('#detail-network-ip').text('')
    $('#detail-nat-ip').text('')
    processBox.innerHTML = '';
    $me.lock();
    $.ajax({
        url: "/admin/api/compute/instances/insert",
        type: 'POST',
        dataType: 'json',
        data: data, 

        error: function(xhr, stateText, err){
            alert(err);
        },
        complete: function(){
            $me.unlock();
        },
    });

});

}).on('/instances.html', function(){

    var collection = {},
        $projectSelect = $('#sel-projects'),
        $zoneSelect = $('#sel-zones'),
        $instances = $('#instances'),
        $tmpInstance = $(`
                        <div class="instance col-xs-6 col-lg-4">
                                <h3></h3> <button class="btn-delete">delete</button>
                                <ul class="inner">
                                    <li class="status">Status<span></span></li>
                                    <li class="ip">Nat IP<span></span></li>
                                    <li class="network_ip">network IP<span></span></li>
                                </ul>
                            </div>
                        `);
    
    connWS(`ws://${location.host}/ws`, function(e){
        
        var data = JSON.parse(e.data),
            project = $projectSelect.val(),
            zone = $zoneSelect.val(),
            m;

        console.log(data)

        if (data.items["project"] != project || data.items["zone"] != zone) {
            return
        }

        if (m = data.active.match(/^compute#instance#([-\w]+)$/)) {

            var name = m[1];
            console.log(name, collection[name])

            if ( ! collection[name]) {
                if (data.items["status"] == "STOPPING") return;
                if (data.items["status"] == "TERMINATED") return;
                $.get(`/admin/api/compute/instance?project=${project}&zone=${zone}&name=${name}`, function(json, stateText, xhr){
                    insert(json);
                });
            }

            if (collection[name]) {
                collection[name].find('.status > span').text(data.items["status"]);
                if (data.items["status"] == "RUNNING") {
                    $.get(`/admin/api/compute/instance?project=${project}&zone=${zone}&name=${name}`, function(json, stateText, xhr){
                        collection[name].data('item', json);
                        displayIP(collection[name], collection[name].data('item'));
                    });
                }
                if (data.items["status"] == "TERMINATED") {
                    collection[name].remove();
                    delete collection[name];
                }
            }

        }
    });

    function displayIP($item, item) {
        var networkIPList = [],
                ipList = [];
        for (var i = 0, iLen = item.networkInterfaces.length; i < iLen; i++) {
                        for (var j = 0, jLen = item.networkInterfaces[i].accessConfigs.length; j < jLen; j++) {
                                            ipList.push(item.networkInterfaces[i].accessConfigs[j].natIP);
                                        }
                        networkIPList.push(item.networkInterfaces[i].networkIP);
                    }
        $item.find('.ip > span').text(ipList.join(','))
        $item.find('.network_ip > span').text(networkIPList.join(','))
        return ipList.pop();
    }

    function insert(item) {
        var $item = $tmpInstance.clone();
        $item.attr("ref", item.name);
        $item.data('item', item);
        $item.find('h3').text(item.name);
        $item.find('.status > span').text(item.status);
        $instances.append($item)
        displayIP($item, item);
        collection[item.name] = $item;
        return $item;
    }

    function initZones(project) {
        $zoneSelect.empty()
        $.get(`/admin/api/compute/zones?project=${project}`, function(json, stateText, xhr){
            $(json.items).each(function(i, item){
                $zoneSelect.append(`<option>${item.name}</option>`);
            });
            $zoneSelect.val('asia-east1-a').change();
        });
    }

    $zoneSelect.on('change', function(){
        $instances.empty();
        var project = $projectSelect.val(),
            zone = $zoneSelect.val();
        // fetch all instances
        $.get(`/admin/api/compute/instances?project=${project}&zone=${zone}`, function(json, stateText, xhr){
            $.each(json.items, function(_, item){
                insert(item);
            }, "json")
        });
    });
    $projectSelect.on('change', function(){
        initZones(this.value)
    }).change();

    // deletion
    $instances.on('click', 'button.btn-delete', function(e){
        var $this = $(this),
        item = $this.parent().data('item');
        if (item.name == 'vm-test-1') {
            alert('這個不能刪')
            return
        }
                                                
        var project = item.zone.replace(/^.*\/projects\/([-\w]+)\/.*$/, '$1'),
            zone = item.zone.replace(/^.*\//, '');

        $this.lock();
                
        $.ajax({
            type: 'DELETE',
            url: `/admin/api/compute/instances/delete?project=${project}&zone=${zone}&name=${item.name}`,
            error: function(){
               $this.unlock();
            },
        });
    });

}).listen();

});
