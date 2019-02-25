{{define "content"}}
<ul id="notifications"></ul>
<div class="main">
    <div class="container">
        <img src="/assets/logo.svg" width="200" height="60" alt="Hiveon ID" class="logo">
        <h1><span class="thin">Recover your </span>Hiveon ID<span class="thin"> password </span></h1>
            <form action="{{mountpathed "recover_end"}}" method="post">
            {{with .errors}}{{with (index . "")}}{{range .}}<span class="err">{{.}}</span>{{end}}{{end}}{{end -}}
            <div class="field">
                    {{with .errors}}{{range .password}}<span class="err">{{.}}</span>{{end}}{{end -}}
                    <input id="password" name="password" type="password" placeholder="Password" />
                    <label for="password">Password</label>
                </div>
                <div class="field">            
                    {{with .errors}}{{range .confirm_password}}<span class="err">{{.}}</span>{{end}}{{end -}}
                    <input id="confirm_password" name="confirm_password" type="password" placeholder="Confirm Password" />
                    <label for="confirm_password">Confirm Password</label>
                </div>
                <input class="submit" type="submit" value="Set password">
            {{with .csrf_token}}<input type="hidden" name="csrf_token" value="{{.}}" />{{end}}
        </form>
    </div>
</div>
{{end}}