{{define "content"}}
<ul id="notifications"></ul>
<div class="main">
    <div class="container">
        <img src="/assets/logo.svg" width="200" height="60" alt="Hiveon ID" class="logo">
        <h1><span class="thin">Set your new password</span></h1>
            <form action="{{mountpathed "recover"}}" method="post">
            {{with .errors}}{{with (index . "")}}{{range .}}<span class="err">{{.}}</span>{{end}}{{end}}{{end -}}
            <div class="field">    
                {{with .errors}}{{range .password}}<span class="err">{{.}}</span>{{end}}{{end -}}
                <input id="email" name="email" type="text" value="{{with .preserve}}{{with .email}}{{.}}{{end}}{{end}}" placeholder="E-mail" />
                <label for="email">E-mail</label>
            </div>                
            <input class="submit" type="submit" value="Reset my password">
            Back to <a href="/login">Login</a>

            {{with .csrf_token}}<input type="hidden" name="csrf_token" value="{{.}}" />{{end}}
        </form>
    </div>
</div>
{{end}}