package web

var TextInputTempalte = `
<div class="form-group">
    <label for="{{.name}}">{{.label}}</label>
    <input type="text" class="form-control" id="{{.name}}" value="">
</div>
`

var ChildCardTemplate = `
<details class="collapse-panel w-400 mw-full mt-10" open>
    <summary class="collapse-header without-arrow">
        <span class="hidden-collapse-open text-muted">&plus;</span>
        <span class="hidden-collapse-closed text-muted">&minus;</span>
        <span class="ml-5">{{.name}}</span>
    </summary>
    <div class="collapse-content">
		{{.content}}
    </div>
</details>
`

var ToggleTemplate = `
<div class="form-group">
    <div class="custom-switch">
        <input type="checkbox" id="{{.name}}" value="" checked="{{.name}}">
        <label for="{{.name}}">{{.label}}</label>
    </div>
</div>
`

var FileTemplate = `
<div class="custom-file">
    <label for="{{.name}}" >{{.label}}</label>
    <input type="file" class="form-control" id="{{.name}}">
</div>
`

var OptionsTemplate = `
<div class="form-group">
    <label for="{{.name}}" >Level</label>
    <select class="form-control" id="{{.name}}">
        {{.options}}
    </select>                                  
</div>
`
var OptionTemplate = `
<option value="" selected="selected" disabled="{{.disabled}}">{{.label}}</option>
`

var ParentCardTemplate = `
<div id="{{.name}}" class="card">
    <div class="position-relative">
        <div class="custom-switch position-absolute bottom-0 bottom-sm-auto top-sm-0 right-0">
            <input type="checkbox" id="{{.name}}Edit" onclick="MakeCurrentFormEditable();" value="">
            <label for="{{.name}}Edit">Edit</label>
        </div>
    </div>
    <h2 class="content-title">
        {{.title}}
    </h2>
    <p>
        {{.content}}
    <div class="custom-switch position-absolute top-0 top-sm-auto bottom-sm-10 right-0">
        <button type="submit" class="btn" id="{{.name}}Fetch" onclick="getSettings();" value="">Fetch</button>
        <button type="submit" class="btn btn-primary" id="{{.name}}Save" onclick="saveSettings();" value="">Save</button>
    </div>
    </p>
</div>
`

var PageTemplate = `
<!DOCTYPE html>
<html>

<head>
    <meta charset='utf-8'>
    <meta http-equiv='X-UA-Compatible' content='IE=edge'>
    <title>AIO Gateway Administration</title>
    <meta name='viewport' content='width=device-width, initial-scale=1'>

    <link href="https://cdn.jsdelivr.net/npm/halfmoon@1.1.1/css/halfmoon-variables.min.css" rel="stylesheet" />
    <script src="https://cdn.jsdelivr.net/npm/halfmoon@1.1.1/js/halfmoon.min.js"></script>

<body class="with-custom-webkit-scrollbars with-custom-css-scrollbars">
    <div id="" class="page-wrapper with-navbar">
        <nav class="navbar">
            <div class="navbar-content">
                    <a href="#" class="navbar-brand">All-in-One Gateway Configuration</a>
            </div>
            <div class="navbar-content ml-auto">
                <div class="custom-switch">
                    <input type="checkbox" id="darkmode" onclick="halfmoon.toggleDarkMode();" value="">
                    <label for="darkmode">Toggle Dark Mode</label>
                </div>
            </div>
        </nav>
        <div class="content-wrapper col-6 offset-3">
         {{.content}}
        </div>
    </div> 
    <script src="https://cdn.jsdelivr.net/npm/jquery@3.6.0/dist/jquery.min.js"></script>
    <script src="admin.js"></script>


</body>

</html>
`
