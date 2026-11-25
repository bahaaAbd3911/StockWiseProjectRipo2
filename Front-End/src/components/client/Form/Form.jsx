// Import Style
import {
  formContainerClass,
  formHeaderClass,
  formTitleClass,
  formWelcommingMsgClass,
  inputsContainerClass,
  buttonContainerClass,
  linksContainerClass,
  imageContainerClass,
  arbFormContainerClass,
  darkFormContainerClass,
  darkLinksContainerClass,
} from "./Sign.module.css";

// Import Hooks
import { useContext, useState } from "react";
import { AuthContext } from "../../../context/AuthContext";
import { useNavigate } from "react-router";

// Import Hooks
import Icon from "../Icon/Icon";
import Input from "./Input";

function Form({
  formTitle,
  arbFormTitle,
  formWelcommingMsg,
  arbFormWelcommingMsg,
  icon,
  status,
  inputsList,
  setSignFormData,
  validationFunction,
  setImage,
  image,
  onSubmit,
}) {
  const [submitting, setSubmitting] = useState(false);

  // Validate first; only valid forms reach onSubmit (which talks to the API).
  const handleSubmit = async () => {
    if (submitting || !validationFunction(setSignFormData, inputsList)) return;
    setSubmitting(true);
    try {
      await onSubmit?.(inputsList);
    } finally {
      setSubmitting(false);
    }
  };

  const navigate = useNavigate();

  const { lang, dark } = useContext(AuthContext);
  return (
    <>
      <div
        className={`${formContainerClass} ${arbFormContainerClass} ${dark && darkFormContainerClass}`}
        onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
      >
        <div className={formHeaderClass}>
          <div className={formTitleClass}>
            {lang === "arb" ? arbFormTitle : formTitle}
          </div>
          <div className={formWelcommingMsgClass}>
            <div>
              <Icon iconSrc={icon} />
            </div>
            <div>
              {lang === "arb" ? arbFormWelcommingMsg : formWelcommingMsg}
            </div>
          </div>
        </div>
        <div className={inputsContainerClass}>
          {inputsList.map((input, index) => (
            <Input
              data={input}
              key={index}
              setSignFormData={setSignFormData}
              setImage={setImage}
              image={image}
            />
          ))}
        </div>
        {status === "Up" && (
          <div className={imageContainerClass}>
            {image && <img src={URL.createObjectURL(image)} alt="" />}
          </div>
        )}
        <div className={buttonContainerClass}>
          <button onClick={handleSubmit} disabled={submitting}>
            {submitting && <i className="fa-solid fa-spinner fa-spin"></i>}{" "}
            {status === "In" && lang === "arb" && "تسجيل الدخول"}
            {status === "In" && lang !== "arb" && "Sign In"}
            {status === "Up" && lang === "arb" && "إنشاء حساب"}
            {status === "Up" && lang !== "arb" && "Sign Up"}
          </button>
        </div>
        <div
          className={`${linksContainerClass} ${dark && darkLinksContainerClass}`}
        >
          <div
            onClick={() => {
              status === "In" ? navigate("/signUp") : navigate("/signIn");
            }}
          >
            {status === "In" && lang === "arb" && "إنشاء حساب جديد"}
            {status === "In" && lang !== "arb" && "Create New Account"}
            {status === "Up" && lang === "arb" && "لدي حساب بافعل"}
            {status === "Up" && lang !== "arb" && "Already Have An Account"}
          </div>
          <div
            onClick={() => {
              navigate("/client/home");
            }}
          >
            {lang === "arb" && "العودة الى الصفحة الرئيسية"}
            {lang !== "arb" && "Go Back To Home"}
          </div>
        </div>
      </div>
    </>
  );
}

export default Form;
