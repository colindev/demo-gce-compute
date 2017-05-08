{{define "content"}}
<div class="row">
    <select id="sel-projects">
        <option>{{ .ProjectID }}</option>
    </select>
    <select id="sel-regions">
        <option value="">global</option>
        <option>asia-east1</option>
    </select>
</div>
<div id="projects" class="row"> <!-- list --> </div>
{{end}}
