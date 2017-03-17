{{define "content"}}
<div class="row">
    <textarea id="startup-script">
#!/usr/bin/env bash

yum update -y
</textarea>
</div>

<div class="row">
    <p class="text-center">
        <button class="btn btn-sm btn-default paging-prev">上一步</button>
        <button class="btn btn-sm btn-primary paging-next">下一步</button>
    </p>
</div>
{{end}}
