import { mainFormContainerClass } from "./Sign.module.css";

import { useState, useContext } from "react";
import { Navigate, useNavigate } from "react-router";

import { AuthContext } from "../../../context/AuthContext";
import { SignUpValidation } from "../../../utils/helpers";
import { readImage, signUpFields, withFieldError } from "../../../data/authForms";
import { splitName } from "../../../api/mappers";
import { useToast } from "../../../hooks/useToast";
import rocketIcon from "../../../assets/images/Icon/rocket.png";

import UserSettings from "../../common/UserSettings/UserSettings";
import Form from "./Form";

const MAX_IMAGE_BYTES = 500 * 1024;

// Public sign-up always creates a customer account; the user lands on the store home page.
function SignUp() {
  const { signUp, isSign } = useContext(AuthContext);
  const navigate = useNavigate();
  const toast = useToast();
  const [fields, setFields] = useState(() => signUpFields());
  const [image, setImage] = useState();

  if (isSign) return <Navigate to="/client/home" replace />;

  const submit = async ([name, email, password, , phone]) => {
    if (image && image.size > MAX_IMAGE_BYTES) {
      setFields((prev) => withFieldError(prev, 5, "Image must be smaller than 500 KB", "يجب أن تكون الصورة أصغر من 500 كيلوبايت"));
      return;
    }

    try {
      const user = await signUp({
        ...splitName(name.data),
        email: email.data.trim(),
        password: password.data,
        phone: String(phone.data).trim(),
        profile_image_url: await readImage(image),
      });
      toast.success(`Welcome to the store, ${user.firstName}!`, `مرحباً بك في المتجر، ${user.firstName}!`);
      navigate("/client/home");
    } catch (err) {
      if (err.status === 409) {
        setFields((prev) => withFieldError(prev, 1, "An account with this email already exists", "يوجد حساب بهذا البريد الإلكتروني"));
      } else {
        toast.error(err.message, err.message);
      }
    }
  };

  return (
    <div className={mainFormContainerClass}>
      <UserSettings />
      <Form
        formTitle={"Sign Up"}
        arbFormTitle={"إنشاء حساب جديد"}
        formWelcommingMsg={"Let's Get Started"}
        arbFormWelcommingMsg={"لنبدأ"}
        icon={rocketIcon}
        inputsList={fields}
        setSignFormData={setFields}
        status={"Up"}
        validationFunction={SignUpValidation}
        setImage={setImage}
        image={image}
        onSubmit={submit}
      />
    </div>
  );
}

export default SignUp;
