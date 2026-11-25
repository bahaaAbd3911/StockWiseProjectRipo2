import { mainFormContainerClass } from "./Sign.module.css";

import { useState, useContext } from "react";
import { Link, Navigate, useNavigate } from "react-router";

import { AuthContext } from "../../../context/AuthContext";
import { SignInValidation } from "../../../utils/helpers";
import { signInFields, withFieldError } from "../../../data/authForms";
import { useToast } from "../../../hooks/useToast";
import HandIcon from "../../../assets/images/Icon/wave.png";

import UserSettings from "../../common/UserSettings/UserSettings";
import Form from "./Form";

function SignIn() {
  const { signIn, isSign, isAdmin, lang } = useContext(AuthContext);
  const navigate = useNavigate();
  const toast = useToast();
  const [fields, setFields] = useState(signInFields);

  if (isSign) return <Navigate to={isAdmin ? "/admin" : "/client/home"} replace />;

  const submit = async ([email, password]) => {
    try {
      const user = await signIn(email.data.trim(), password.data);
      toast.success(`Welcome back, ${user.firstName}!`, `أهلاً بعودتك، ${user.firstName}!`);
      navigate(user.role === "admin" ? "/admin" : "/client/home");
    } catch (err) {
      if (err.status === 401) {
        setFields((prev) => withFieldError(prev, 1, "Invalid email or password", "البريد الإلكتروني أو كلمة المرور غير صحيحة"));
      } else {
        toast.error(err.message, err.message);
      }
    }
  };

  return (
    <div className={mainFormContainerClass}>
      <UserSettings />
      <Form
        formTitle={"Sign In"}
        arbFormTitle={"تسجيل الدخول"}
        formWelcommingMsg={"Welcome Again"}
        arbFormWelcommingMsg={"أهلا بك من جديد"}
        icon={HandIcon}
        inputsList={fields}
        setSignFormData={setFields}
        status={"In"}
        validationFunction={SignInValidation}
        onSubmit={submit}
      />
      <Link to="/reset-password" style={{ marginTop: 12, color: "var(--color-primary)" }}>
        {lang === "arb" ? "نسيت كلمة المرور؟" : "Forgot your password?"}
      </Link>
    </div>
  );
}

export default SignIn;
