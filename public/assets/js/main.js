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
    $(this._pages).each(function(i, path){
        if (path == location.pathname) {
            cur = i;
        }
    });
    this.current = cur;
}

Paging.prototype = {
    next: function(){
        location.href = this._pages[this.current+1];
    },
    prev: function(){
        if (this.current) {
            location.href = this._pages[this.current-1];
        }
    },
    bind: function(prev, next){
        var me = this;
        $(prev).on('click', function(e){
            me.prev();
        });
        $(next).on('click', function(e){
            me.next();
        });

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
}
$.fn.radioButtonBox = function(prefix, conf){
    var box = this,
        propName = prefix.replace(/[^a-z]*$/, ''),
        rePrefix = new RegExp(`^${prefix}`),
        collect = [];

    $(box).find(`[id^=${prefix}]`).each(function(){
        
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
        image: "centos-7-v20170227",
        layout: "1",
        cpu: "1",
        memory: "1024"
    }), 
    page = (new Paging([
        "/", 
        "/machine_type.html", 
        "/startup_script.html",
        "/create.html"])).bind('button.paging-prev', 'button.paging-next');

$('[id^=layout-]').on('click', function(e){
    
    conf.set('layout', this.id.replace(/^layout-/, '')).store();
    page.next();

}).each(function(){
    if (this.id == 'layout-'+conf.get('layout')) {
        $(this).active();
        return
    }

    $(this).unactive();
});
$('#cpu').radioButtonBox('cpu-', conf);
$('#memory').radioButtonBox('memory-', conf);
$('#image-name').text(conf.get('image'));

$('#btn-create').on('click', function(){

    var data = JSON.parse(localStorage.getItem('config'));

    console.log(data)
    return
    // TODO 
    // confirm instance
    // lock button
    // unlock button

    $.ajax({
        type: 'POST',
        dataType: 'json',
        data: data, 

        error: function(){},
        complete: function(){
            // unlock button
        },
    });

});

})(jQuery)
