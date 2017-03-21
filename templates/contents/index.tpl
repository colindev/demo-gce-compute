{{define "content"}}
<h3>選擇 CPU 數量</h3>
<div id="cpu" class="row">
    <div class="col-md-3">
        <button id="cpu-1">CPU 1</button>
    </div>
    <div class="col-md-3">
        <button id="cpu-2">CPU 2</button>
    </div>
    <div class="col-md-3">
        <button id="cpu-4">CPU 4</button>
    </div>
    <div class="col-md-3">
        <button id="cpu-6">CPU 6</button>
    </div>
    <div class="col-md-3">
        <button id="cpu-8">CPU 8</button>
    </div>
    <div class="col-md-3">
        <button id="cpu-10">CPU 10</button>
    </div>
</div>

<h3>選擇 記憶體大小</h3>
<div class="row">
    <input id="display-memory"/>MB
    <input type="range" id="memory" min="1024" max="4096" step="64">
</div>

<div class="row">
    <p class="text-center">
        <button class="btn btn-sm btn-default paging-prev">上一步</button>
        <button class="btn btn-sm btn-primary paging-next">下一步</button>
    </p>
</div>
{{end}}
