(function(factory){
    self["Metadata"] = factory(self.JSON);
})(function(JSON){

function Metadata(key, o) {

    this.key = key
    this._default = o || {};
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
    _default: {},
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
    getDefault: function(name){
        return this._default[name];
    }
};

return Metadata;

});
