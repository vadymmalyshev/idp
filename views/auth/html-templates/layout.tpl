<html>
    <head>
        <meta charset="utf-8">
        <meta http-equiv="x-ua-compatible" content="ie=edge">
        <meta name="application-name" content="Hiveon ID">
        <meta name="msapplication-TileColor" content="#1f2228">
        <meta name="theme-color" content="#1f2228">
        <link rel="safari-pinned-tab" href="/assets/safari-pinned-tab.svg">
        <meta name="apple-mobile-web-app-title" content="Hiveon ID">
        <title>{{template "pagetitle" .}}</title>
        <meta name="description" content="">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <script charset="utf-8" src="/assets/jquery.min.js"></script>
        <link rel="preload" href="/assets/logo.svg" as="image">
        <link type="image/x-icon" rel="icon" href="/assets/favicon-32x32.png">
        <link rel="stylesheet" href="/assets/styles.css" class="rel">
        <link rel="stylesheet" href="/fonts/Gilroy/gilroy.css" class="rel">
        <link rel="stylesheet" href="/fonts/OpenSans/opensans.css" class="rel">
    </head>
<body>
<script charset="utf-8" src="/assets/scripts.js"></script>
	{{with .flash_success}}<div class="alert alert-success">{{.}}</div>{{end}}
	{{with .flash_error}}<div class="alert alert-danger">{{.}}</div>{{end}}
	{{template "content" .}}

{{with .csrf_token}}<input type="hidden" name="csrf_token" value="{{.}}" />{{end}}
</body>
</html>
{{define "pagetitle"}}{{end}}
{{define "content"}}{{end}}