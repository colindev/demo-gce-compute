(function(factory){

    self["Paging"] = factory(self.jQuery)

})(function($){

 
function Paging(arr) {
    this._pages = arr;
    this._handlers = {};

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
    },
    on: function(path, handler){
        if (this._handlers[path]) {
            throw new Error("already exists: "+path);
        }

        if ( ! $.isArray(path)) {
            path = [path];
        }

        for (var i=0,L=path.length; i<L; i++) {
            this._handlers[path[i]] = handler;
        }
        
        return this;
    },
    trigger: function(path){
        var handler = this._handlers[path];
        if ( ! handler) {
            return
        }

        handler.call(this);
    },
    listen: function(){
        this.trigger(location.pathname);
    }
};

return Paging

});
