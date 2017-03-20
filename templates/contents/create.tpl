{{define "content"}}
<div class="row">
    <ul>
        <li>虛擬機名稱:<span id="detail-name"></span></li>
        <li>CPU:<span id="detail-cpu"></span></li>
        <li>記憶體:<span id="detail-memory"></span></li>
        <li>內網IP:<span id="detail-network-ip"></span></li>
        <li>外網IP:<span id="detail-nat-ip"></span></li>
    </ul>
</div>

<div class="row">
    <p class="text-center">
        <button class="btn btn-sm btn-default" onclick="location.href='/'">建立新虛擬機</button>
        <button class="btn btn-sm btn-default paging-prev">上一步</button>
        <button class="btn btn-sm btn-default paging-next">下一步</button>
        <button id="btn-create" class="btn btn-sm btn-primary">建立</button>
    </p>
</div>

<div class="row">
    <div id="process-status"></div>
</div>

{{end}}
