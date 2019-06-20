import { h, Component } from 'preact';
import { Link } from 'preact-router/match';
// import { fetch } from 'whatwg-fetch';
import fetch from 'unfetch';
import eyeImg from '../../assets/eye.svg';
import ButtonLoader from '../../components/buttonLoader';
import FormInput from '../../components/formInput';
import API from '../../config/api';
import logo from '../../assets/logo.svg';
import style from './style.css';

export default class Login extends Component {
	state = {
    login: null,
    shouldShowPass: false,
    count: 10,
    errors: {},
    touched: {},
    isSubmitting: false,
    remember: false,
	};


	// gets called when this route is navigated to
	componentDidMount() {
		// start a timer for the clock:
	}

	// gets called just before navigating away from the route
	componentWillUnmount() {
  }
  
  handleSubmit = e => {
    e.preventDefault();
    e.stopPropagation();
    // debugger;
    const {
      login, password
    } = this.state;
    fetch(API.login, {
      method: 'POST',
      body: JSON.stringify({
        email: login, password
      }),
      // mode: 'cors',
      // credentials: 'include',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        // 'Origin': 'https://id.hiveon.dev',
        // 'Refferer': 'https://id.hiveon.dev/login',
        // 'Host': 'https://id.hiveon.dev',
        // ...(cookie ? { Cookie: cookie } : null),
      },
    });
    return false;
  };

  errorsChecker = {
    login: {
      valueMissing: 'Required login or email',
    },
    password: {
      valueMissing: 'Password required',
    },
  };

  validateField = (name, validity) => {
    const errorField = this.errorsChecker[name];
    if(!errorField){
      return false;
    }
    const {
      badInput,
      customError,
      patternMismatch,
      rangeOverflow,
      rangeUnderflow,
      stepMismatch,
      tooLong,
      tooShort,
      typeMismatch,
      valid,
      valueMissing,
    } = validity || {};
    
    if(valueMissing && errorField.valueMissing) {
      return errorField.valueMissing;
    }

    return false;
  };

  handleChange = e => {
    const {name, value, validity } = e.target;
    console.log('validity', validity)
    
    this.setState({
      [name]: value, 
      errors: {
        ...this.state.errors,
        [name]: this.validateField(name, validity),
      },
      touched: {
        ...this.state.touched,
        [name]: true,
      }
    });
  };

  handleBlur = () => {

  };

  togglePasswordInputType = () => {
    this.setState(({shouldShowPass}) => ({
      shouldShowPass: !shouldShowPass
    }))
  };

	// Note: `user` comes from the URL, courtesy of our router
	render({ user }, { remember, code, login, password, errors, touched, shouldShowPass, isSubmitting }) {
		return (
			<div class={style.login}>
        <div className={style.container}>
        <div className={style.logo}>
          <img src={logo} alt="HiveonID"/>
        </div>
        <div className={style.title}>
          Log in your account
        </div>
				
        <form method="post" onSubmit={this.handleSubmit}>
          <div className={style.fieldset}>
              <FormInput
                className={style.input}
                type="text"
                name="login"
                placeholder="Login or Email"
                value={login}
                onChange={this.handleChange}
                autoComplete="username"
                error={errors.login}
                required
                autoFocus // eslint-disable-line jsx-a11y/no-autofocus
              />
              
            </div>
            <div className={style.fieldset}>
              <div className={style.inputContainer}>
                <FormInput
                  className={style.input}
                  type={shouldShowPass ? 'text' : 'password'}
                  name="password"
                  placeholder="Password"
                  value={password}
                  onChange={this.handleChange}
                  autoComplete="current-password"
                  error={errors.password}
                  required
                />
                  <img onClick={this.togglePasswordInputType} src={eyeImg} className={style.eye} alt="" />
                </div>
            </div>
            <div className={style.fieldset}>
              <div className={style.inputContainer}>
                  <FormInput
                    className={style.input}
                    type="text"
                    name="twofa_code"
                    placeholder="2FA code (if enabled)"
                    value={code}
                    onChange={this.handleChange}
                    error={errors.twofa_code}
                    autoComplete="off"
                  />
                </div>
                {/* <div className={style.errorMessage}>
                  {errors && errors.password && touched.password
                    ? "Not correct 2fa code"
                    : ''}
                </div> */}
            </div>

            <div className={style.flexSpaceBetween}>
              <label className={style.checkboxLabel}>
                <input
                  name="remember"
                  type="checkbox"
                  checked={remember}
                  value="1"
                  className={style.checkbox}
                  onChange={this.handleChange}
                />
                <span className={style.checkboxLabelText}>
                  Remember me
                </span>
              </label>
              <Link className={style.link} href="/restore-pass/">
                Forgot password?
              </Link>
            </div>

            <button className={style.button} type="submit" disabled={isSubmitting}>
              <ButtonLoader isLoading={isSubmitting}>
                Log in
              </ButtonLoader>
            </button>

            <div className={style.bottomRelocateBlock}>
              <span>Don't have account yet?</span>{' '}
              <Link className={style.link} href="/register/">
                Register
              </Link>
            </div>
          </form>
        </div>
			</div>
		);
	}
}
