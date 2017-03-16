(function($){

function Config(o) {

    try {
        o = JSON.parse(localStorage.getItem('config'));
    } catch(e) {} 
    for (var n in o) {
        this.set(n, o[n])
    }
}
Config.prototype = {
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
        localStorage.setItem("config", JSON.stringify(this.obj()));

        console.log(this.obj())
    },
};

$.fn.active = function(){
    this.addClass('active');
};
$.fn.unactive = function(){
    this.removeClass('active');
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

};

var conf = new Config({
    layout: "1",
    cpu: "1",
    memory: "1024"
}); 

$('[id^=layout-]').on('click', function(e){
    conf.set('layout', this.id.replace(/^layout-/, '')).store();
    location.href = '/machine_type.html'
}).each(function(){
    if (this.id == 'layout-'+conf.get('layout')) {
        $(this).active();
        return
    }

    $(this).unactive();
});
$('#cpu').radioButtonBox('cpu-', conf);
$('#memory').radioButtonBox('memory-', conf);
$('#btn-machine-type-prev').on('click', function(e){
    location.href = '/';
});
$('#btn-machine-type-next').on('click', function(e){
    location.href = '/install-app.html';
});

})(jQuery)
