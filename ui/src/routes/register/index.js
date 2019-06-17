import { h, Component } from 'preact';
import { Link } from 'preact-router/match';
import eyeImg from '../../assets/eye.svg';
import ButtonLoader from '../../components/buttonLoader';
import logo from '../../assets/logo.svg';
import style from './style';

export default class Login extends Component {
	state = {
    name: '',
    login: '',
    email: '',
    password: '',
    confirm_password: '',
    shouldShowPass: false,
    count: 10,
    errors: {},
    touched: {},
    terms: true,
    isSubmitting: false,
	};


	// gets called when this route is navigated to
	componentDidMount() {
		// start a timer for the clock:
	}

	// gets called just before navigating away from the route
	componentWillUnmount() {
  }
  
  handleSubmit = () => {

  };

  handleChange = e => {
    const {name, value} = e.target;
    this.setState({[name]: value, touched: {
      ...this.state.touched,
      [name]: true,
    }});
  };

  handleBlur = () => {

  };

  togglePasswordInputType = () => {
    this.setState(({shouldShowPass}) => ({
      shouldShowPass: !shouldShowPass
    }))
  };

	// Note: `user` comes from the URL, courtesy of our router
	render({ user }, { 
    name,
    code, 
    login,
    email, 
    password,
    confirm_password: confirmPassword,
    errors, 
    terms,
    shouldShowPass, 
    isSubmitting }) {
		return (
			<div class={style.login}>
        <div className={style.container}>
          <div className={style.logo}>
            <img src={logo} alt="HiveonID"/>
          </div>
          <div className={style.title}>
            Create new account
          </div>
				
          <form method="post" onSubmit={this.handleSubmit}>
            <div className={style.fieldset}>
              <input
                className={style.input}
                type="text"
                name="name"
                placeholder="Name"
                value={name}
                onChange={this.handleChange}
                onBlur={this.handleBlur}
                autoComplete="name"
                ref={this.refInputLogin}
                autoFocus // eslint-disable-line jsx-a11y/no-autofocus
              />
              <div className={style.errorMessage}>
                {errors && errors.login && touched.login
                  ? 'Name Required'
                  : ''}
              </div>
            </div>
            <div className={style.fieldset}>
                <input
                  className={style.input}
                  type="text"
                  name="login"
                  placeholder="Login"
                  value={login}
                  onChange={this.handleChange}
                  onBlur={this.handleBlur}
                  autoComplete="username"
                  ref={this.refInputLogin}
                  autoFocus // eslint-disable-line jsx-a11y/no-autofocus
                />
                <div className={style.errorMessage}>
                  {errors && errors.login && touched.login
                    ? 'Required logins'
                    : ''}
                </div>
              </div>
              <div className={style.fieldset}>
                <input
                  className={style.input}
                  type="email"
                  name="email"
                  placeholder="Email"
                  value={email}
                  onChange={this.handleChange}
                  onBlur={this.handleBlur}
                  autoComplete="email"
                  ref={this.refInputLogin}
                  autoFocus // eslint-disable-line jsx-a11y/no-autofocus
                />
                <div className={style.errorMessage}>
                  {errors && errors.login && touched.login
                    ? 'Required logins'
                    : ''}
                </div>
              </div>
              <div className={style.fieldset}>
                <div className={style.inputContainer}>
                    <input
                      className={style.input}
                      type={shouldShowPass ? 'text' : 'password'}
                      name="password"
                      placeholder="Password"
                      value={password}
                      onChange={this.handleChange}
                      onBlur={this.handleBlur}
                      autoComplete="current-password"
                      ref={this.refInputPassword}
                    />
                    <img onClick={this.togglePasswordInputType} src={eyeImg} className={style.eye} alt="" />
                  </div>
                  <div className={style.errorMessage}>
                    {errors && errors.password && touched.password
                      ? 'Password required'
                      : ''}
                  </div>
              </div>
              <div className={style.fieldset}>
                <div className={style.inputContainer}>
                    <input
                      className={style.input}
                      type={shouldShowPass ? 'text' : 'password'}
                      name="confirm_password"
                      placeholder="Confirm Password"
                      value={confirmPassword}
                      onChange={this.handleChange}
                      onBlur={this.handleBlur}
                      autoComplete="current-password"
                      ref={this.refInputPassword}
                    />
                    <img onClick={this.togglePasswordInputType} src={eyeImg} className={style.eye} alt="" />
                  </div>
                  <div className={style.errorMessage}>
                    {errors && errors.confirm_password && touched.confirm_password
                      ? 'Passwords do not match'
                      : ''}
                  </div>
              </div>
              <div className={style.fieldset}>
                <div className={style.inputContainer}>
                    <input
                      className={style.input}
                      type="text"
                      name="promocode"
                      placeholder="Promocode"
                      value={code}
                      onChange={this.handleChange}
                      autoComplete="off"
                      onBlur={this.handleBlur}
                    />
                  </div>
                  <div className={style.errorMessage}>
                    {errors && errors.password && touched.password
                      ? "Not correct 2fa code"
                      : ''}
                  </div>
              </div>

              <div className={style.flexSpaceBetween}>
                <label className={style.checkboxLabel}>
                  <input
                    name="remember"
                    type="checkbox"
                    checked={terms}
                    value="1"
                    className={style.checkbox}
                    onChange={this.handleChange}
                  />
                  <span className={style.checkboxLabelText}>
                    By clicking register you agree with{' '}
                    <Link className={style.link} href="/restore-pass/">
                      Terms
                    </Link>
                  </span>
                </label>
              </div>

              <button className={style.button} type="submit" disabled={isSubmitting}>
                <ButtonLoader isLoading={isSubmitting}>
                  Register
                </ButtonLoader>
              </button>

              <div className={style.bottomRelocateBlock}>
                <span>Have an account already?</span>{' '}
                <Link className={style.link} href="/login/">
                  Log In
                </Link>
              </div>
            </form>
        </div>
			</div>
		);
	}
}
