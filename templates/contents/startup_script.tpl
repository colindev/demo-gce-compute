{{define "content"}}

<div class="row">
    Hostname: <input id="compute-name"> 
    <p>(名稱開頭必須為小寫字母，後方最多可接 63 個小寫字母、數字或連字號，但結尾不得為連字號)</p>
</div>

<div class="row">
    <textarea id="startup-script"></textarea>
</div>

<div class="row">
    <p class="text-center">
        <button class="btn btn-sm btn-default paging-prev">上一步</button>
        <button class="btn btn-sm btn-primary paging-next">下一步</button>
    </p>
</div>

{{end}}
