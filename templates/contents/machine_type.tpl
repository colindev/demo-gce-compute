{{define "content"}}
<h3>選擇 CPU 數量</h3>
<div id="cpu" class="row">
    <div class="col-md-4">
        <button id="cpu-1">CPU 1</button>
    </div>
    <div class="col-md-4">
        <button id="cpu-2">CPU 2</button>
    </div>
    <div class="col-md-4">
        <button id="cpu-4">CPU 4</button>
    </div>
</div>

<h3>選擇 記憶體大小</h3>
<div id="memory" class="row">
    <div class="col-md-4">
        <button id="memory-512">Memory 512</button>
    </div>
    <div class="col-md-4">
        <button id="memory-1024">Memory 1024</button>
    </div>
    <div class="col-md-4">
        <button id="memory-2048">Memory 2048</button>
    </div>
</div>

<div class="row">
    <p class="text-center">
        <button class="btn btn-sm btn-default paging-prev">上一步</button>
        <button class="btn btn-sm btn-primary paging-next">下一步</button>
    </p>
</div>
{{end}}
