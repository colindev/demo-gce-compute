{{define "content"}}
<div class="row">
    <select id="sel-projects">
        <option>{{ .ProjectID }}</option>
    </select>
    <select id="sel-zones">
    </select>
</div>
<div id="instances" class="row"> <!-- list instances --> </div>
{{end}}
