function togglePasswordView(e, eye) {
    $(this).toggleClass('opened');

    var input = $(this).siblings('input');

    if (input.attr("type") === "password") {
        input.attr("type", "text");
    } else {
        input.attr("type", "password");
    }
}

function resolveTimeZoneOptions() {
    var timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    var hrs = -(new Date().getTimezoneOffset() / 60);

    $('input[name=timezone]') = timezone;
    $('input[name=utc]') = hrs;
}

$(function () {
    $('.eye').click(togglePasswordView);

    $('#login-btn').click(doLogin);
    $('#code-btn').click(doRegister);
    $('#register-btn').click(doCheckToken);

    resolveTimeZoneOptions();
});

function validatePassword() {
    let match = $("#password").val() === $("#confirm-password").val()
    if (!match) {
        alert("Passwords do not match!")
    }
    return match
}

function doLogin() {
    var formData = $('form').serialize();
    
    $.post("/login", formData).done(function (data) {
        $('#notifications').html($('<li>').attr('class', 'ok').text("Signed in"));
        window.location.href = data.redirect_url;
    }).fail(function (data) {
        $('#notifications').html($('<li>').attr('class', 'error').text(data.responseJSON.error));
    });
}

function checkRegistrationForm() {
    var result = true;
    $('#notifications').html('');

    var isUsernameOK = $('input[name=username]').val().length > 5;
    if(!isUsernameOK) {
        result = false;
        $('#notifications').append($('<li>').attr('class', 'error').text("Username length must be > 5 characters"));
    }

    var emailRegex = /^([a-zA-Z0-9_.+-])+\@(([a-zA-Z0-9-])+\.)+([a-zA-Z0-9]{2,4})+$/;

    var isEmailOK = emailRegex.test($('input[name=email]').val());
    if(!isEmailOK) {
        result = false;
        $('#notifications').append($('<li>').attr('class', 'error').text("Email must be valid"));
    }

    // var isPasswordOK = $('input[name=password]').val().length > 5;
    // if(!isPasswordOK) {
    //     result = false;
    //     $('#notifications').append($('<li>').attr('class', 'error').text("Password length must be > 5 characters"));
    // }

    if(!result) {
        $('html, body').animate({ scrollTop: $('#notifications').offset().top }, 'slow');
    }

    return result;
}

function doRegister(e) {
    var isFormOK = checkRegistrationForm();
    if(!isFormOK) {
        return;
    }

    var formData = $('form').serialize();
    $.post("/register", formData).done(function (data) {

    }).done(function (data) {
        if (data.wait_for_code) {
            disabledPreConfirmationFields();
            
            $('#code-btn').addClass('hide');
            $('.sent').removeClass('hide');
            
            $('#register-btn').prop('disabled', 'false');
            $('#register-btn').removeClass('disabled');
        }
    }).fail(function (data) {
        $('#notifications').html($('<li>').attr('class', 'error').text(data.responseJSON.error));
    });
}

function disabledPreConfirmationFields() {
    $('input[name=username]').prop('disabled', true);
    $('input[name=email]').prop('disabled', true);
    $('input[name=password]').prop('disabled', true);
}

function doCheckToken(e) {
    var code = $('input[name=code]').val();
    var email = $('input[name=email]').val();

    $.post('/confirmation', {code: code, email: email})
        .done(function (data) {
            const params = new URLSearchParams(window.location.search);
            const code = params.get('login_challenge') && params.get('login_challenge').replace(/\s+/g,"") || '';

            window.location.href=window.location.origin+"/login?login_challenge=" + code;            
        })
        .fail(function(data){
            if (!data.active) {
                $('#notifications').html($('<li>').attr('class', 'error').text(data.responseJSON.error));
            }
        });
}