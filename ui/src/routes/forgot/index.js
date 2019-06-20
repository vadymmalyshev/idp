import { h, Component } from 'preact';
import { Link } from 'preact-router/match';
import ButtonLoader from '../../components/buttonLoader';
import logo from '../../assets/logo.svg';
import style from './style';

export default class Forgot extends Component {
	state = {
    login: '',
    isSendPass: false,
    count: 10,
    errors: {},
    touched: {},
    isSubmitting: false,
	};

	// update the current time
	// gets called when this route is navigated to
	componentDidMount() {
		// start a timer for the clock:
	}

	// gets called just before navigating away from the route
	componentWillUnmount() {
    this.setState({isSendPass: false});
  }
  
  handleSubmit = () => {

  };

  handleChange = () => {

  };

  handleBlur = () => {

  };

  togglePasswordInputType = () => {
    this.setState(({shouldShowPass}) => ({
      shouldShowPass: !shouldShowPass
    }))
  };

	// Note: `user` comes from the URL, courtesy of our router
	render({ user }, { login, errors, isSendPass, isSubmitting }) {
		return (
			<div class={style.login}>
        <div className={style.container}>
        <div className={style.logo}>
          <img src={logo} alt="HiveonID"/>
        </div>
        <div className={style.title}>
          Enter email to recover password
        </div>
				
        <form method="post" onSubmit={this.handleSubmit}>
          {
            !isSendPass && (
              <div className={style.fieldset}>
                <input
                  className={style.input}
                  type="text"
                  name="login"
                  placeholder="Login or Email"
                  value={login}
                  onChange={this.handleChange}
                  onBlur={this.handleBlur}
                  autoComplete="username"
                  ref={this.refInputLogin}
                  autoFocus // eslint-disable-line jsx-a11y/no-autofocus
                />
                <div className={style.errorMessage}>
                  {errors && errors.login && touched.login
                    ? 'Required login or email'
                    : ''}
                </div>
              </div>
            )
          } 
          {
            !isSendPass && (
              <button className={style.button} type="submit" disabled={isSubmitting}>
                <ButtonLoader isLoading={isSubmitting}>
                  Reset Password
                </ButtonLoader>
              </button>
            )
          }
          {
            isSendPass && (
              <div className={style.sendPass}>
                Password successfully sent, please check your email
              </div>
            )
          }

            <div className={style.bottomRelocateBlock}>
              <span>Return to</span>{' '}
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
