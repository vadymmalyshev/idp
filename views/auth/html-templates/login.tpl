{{define "content"}}
<ul id="notifications"></ul>
<div class="main">
    <div class="container">
        <img src="/assets/logo.svg" width="200" height="60" alt="Hiveon ID" class="logo">
        <h1><span class="thin">Log in with your </span>Hiveon ID</h1>
            <form action="{{mountpathed "login"}}" method="post">
            {{with .errors}}{{with (index . "")}}{{range .}}<span>{{.}}</span><br />{{end}}{{end}}{{end -}}
            <label for="email">E-mail:</label>
            <input name="email" type="text" value="{{with .preserve}}{{with .email}}{{.}}{{end}}{{end}}" placeholder="E-mail" /><br />
            {{with .errors}}{{range .email}}<span>{{.}}</span><br />{{end}}{{end -}}
            <label for="password">Password:</label>
            <input name="password" type="password" placeholder="Password" /><br />
            {{with .errors}}{{range .password}}<span>{{.}}</span><br />{{end}}{{end -}}
            <label><input type="checkbox" name="remember">Remember me</label>
            <input type="submit" value="Register"><br />
            <a href="/">Cancel</a>

            {{with .csrf_token}}<input type="hidden" name="csrf_token" value="{{.}}" />{{end}}
        </form>

    </div>
</div>
{{end}}