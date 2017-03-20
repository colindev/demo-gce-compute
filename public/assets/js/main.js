(function($){

function Metadata(key, o) {

    this.key = key
    try {
        x = JSON.parse(localStorage.getItem(this.key));
        for (var n in o) {
            if (!x.hasOwnProperty(n)) {
                x[n] = o[n];
            }
        }
        o = x
    } catch(e) {} 
    for (var n in o) {
        this.set(n, o[n])
    }
    this.store()
}
Metadata.prototype = {
    _o: {}, 
    set: function(name, value){
        this._o[name] = value;
        return this;
    },
    get: function(name){
        return this._o[name];
    },
    obj: function(){
        return $.extend({}, this._o);
    },
    store: function(){
        localStorage.setItem(this.key, JSON.stringify(this.obj()));

        console.log(this.obj())

        return this;
    },
};

function Paging(arr) {
    this._pages = arr;
    var cur = 0;
    $(this._pages).each(function(i, item){
        if (item.path == location.pathname) {
            cur = i;
        }
    });
    this._current = cur;
}

Paging.prototype = {
    next: function(){
        location.href = this._pages[this._current+1].path;
    },
    prev: function(){
        if (this.current) {
            location.href = this._pages[this._current-1].path;
        }
    },
    current: function(prop){
        return this._pages[this._current][prop];
    },
    bind: function(prev, next){
        var me = this;
        $(prev).on('click', function(e){
            me.prev();
        });
        $(next).on('click', function(e){
            me.next();
        });

        if (this._current == 0) {
            $(prev).hide();
        }

        if (this._current == this._pages.length -1) {
            $(next).hide();
        }

        return this;
    }
};

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
    }
};
$.fn.unlock = function(){
    if (this.prop('tagName').match(/^(input|button)$/i)) {
        this.removeAttr('readonly');
    }
};
$.fn.radioButtonBox = function(prefix, conf){
    var box = this,
        propName = prefix.replace(/[^a-z]*$/, ''),
        rePrefix = new RegExp(`^${prefix}`),
        collect = [];

    $(box).find(`button[id^=${prefix}]`).each(function(){
        
        collect.push(this);
        if (this.id == prefix+conf.get(propName)) {
            $(this).active()
            return
        }
        $(this).unactive();

    }).on('click', function(e){
        
        var me = this;
        conf.set(propName, this.id.replace(rePrefix, '')).store();
        $(collect).each(function(i, btn){
            if (btn === me) {
                $(this).active();
                return
            }
            $(this).unactive();
        });
    });

    return this;
};

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

// ---
$('#cpu').radioButtonBox('cpu-', conf);
$('#memory').radioButtonBox('memory-', conf);

// ---
$('#startup-script').val(conf.get("startup_script")).on('change', function(e){
    conf.set("startup_script", this.value).store();
});

// ---
$('input#compute-name').on('change', function(){
    conf.set('name', this.value).store();
}).val(conf.get('name'));

// ---
var ws = null,
    processBox = document.getElementById('process-status');

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
function connWS(url) {
    ws = new WebSocket(url);
    ws.onmessage = onmessage;
    ws.onopen = function(){console.log('connected')};
    ws.onclose = function(){
        setTimeout(function(){console.log('try reconnect');connWS(url)}, 3000);
    };
}

connWS(`ws://${location.host}/ws`)

$('#detail-cpu').text(conf.get('cpu'));
$('#detail-memory').text(conf.get('memory'));
$('#detail-name').text(conf.get('name'));
$('#btn-create').on('click', function(){

    var $me = $(this),
        data = JSON.parse(localStorage.getItem('config'));

    console.log(data)
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


})(jQuery)
