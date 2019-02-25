{{define "content"}}
<ul id="notifications"></ul>
<div class="main">
    <div class="container">
        <img src="/assets/logo.svg" width="200" height="60" alt="Hiveon ID" class="logo">
        <h1><span class="thin">Register your </span>Hiveon ID</h1>
            <form action="{{mountpathed "register"}}" method="post">
            {{with .errors}}{{with (index . "")}}{{range .}}<span>{{.}}</span><br />{{end}}{{end}}{{end -}}
            <div class="field">
                {{with .errors}}{{range .name}}<span class="err">{{.}}</span>{{end}}{{end -}}
                <input id="name" name="name" type="text" value="{{with .preserve}}{{with .name}}{{.}}{{end}}{{end}}" placeholder="Your Name" />
                <label for="name">Name</label>
            </div>
            <div class="field">            
                {{with .errors}}{{range .email}}<span class="err">{{.}}</span>{{end}}{{end -}}
                <input  id="email" name="email"type="text" value="{{with .preserve}}{{with .email}}{{.}}{{end}}{{end}}" placeholder="E-mail" />
                <label for="email">E-mail</label>
            </div>
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
            <input type="submit" class="submit" value="Register"><br />
            Have an account already? <a href="/login">Log in</a>
            {{with .csrf_token}}<input type="hidden" name="csrf_token" value="{{.}}" />{{end}}
        </form>
    </div>
</div>
{{end}}