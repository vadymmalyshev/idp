{{define "content"}}
<ul id="notifications"></ul>
<div class="main">
    <div class="container">
        <img src="/assets/logo.svg" width="200" height="60" alt="Hiveon ID" class="logo">
        <h1><span class="thin">Register your </span>Hiveon ID</h1>
            <form action="{{mountpathed "login"}}" method="post">
            {{with .errors}}{{with (index . "")}}{{range .}}<span class="err">{{.}}</span>{{end}}{{end}}{{end -}}
            <div class="field">
                {{with .errors}}{{range .email}}<span class="err">{{.}}</span>{{end}}{{end -}}
                <input id="email" name="email" type="text" value="{{with .preserve}}{{with .email}}{{.}}{{end}}{{end}}" placeholder="E-mail" />
                <label for="email">E-mail</label>    
            </div>                
            <div class="field">
                {{with .errors}}{{range .password}}<span class="err">{{.}}</span>{{end}}{{end -}}
                <input id="password" name="password" type="password" placeholder="Password" />
                <label for="password">Password</label>
            </div>
            <div class="in-row">
                <label style="display: flex"><input type="checkbox" name="remember">Remember me</label> <a href="/recover">Forgot password?</a>
            </div>            
            <input class="submit" type="submit" value="Sign In">
            Don't have an account yet? <a href="/register">Register</a>
            {{with .csrf_token}}<input type="hidden" name="csrf_token" value="{{.}}" />{{end}}
        </form>
    </div>
</div>
{{end}}