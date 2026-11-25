import { useContext } from "react";

import {
  subFormClass,
  FormInputContainerClass,
  formInputsContainerClass,
  darkSubFormClass,
  errorMsgClass,
  beforeLastClass,
} from "./ShippingForm.module.css";
import { jordanianAreasList, jordanianCitiesList } from "../../../../data/jordanLocations";
import { AuthContext } from "../../../../context/AuthContext";
import location from "../../../../assets/images/Icon/location.png";
import FormHeader from "../FormHeader";
import ChoicePicker from "./ChoicePicker";

const CITY = 3;
const AREA = 4;
const ADDRESS = 5;

// Delivery address: city -> area (filtered by city) -> address details.
function ShippingFormForm2() {
  const { lang, setorderFormFields, orderFormFields } = useContext(AuthContext);
  const isAr = lang === "arb";
  const error = (i) => (isAr ? orderFormFields[i].arbErrorMsg : orderFormFields[i].errorMsg);

  const setField = (index, value, extra = {}) =>
    setorderFormFields((prev) =>
      prev.map((f, i) =>
        i === index ? { ...f, data: value, flag: false, errorMsg: "", arbErrorMsg: "" } : extra[i] !== undefined ? { ...f, data: extra[i] } : f,
      ),
    );

  const city = orderFormFields[CITY].data;
  const cities = jordanianCitiesList.map((c) => ({ value: c.cityName, label: isAr ? c.arbCityName : c.cityName }));
  const areas = jordanianAreasList
    .filter((a) => a.areaCity === city)
    .map((a) => ({ value: a.areaName, label: isAr ? a.arbAreaName : a.areaName }));

  return (
    <div className={`${subFormClass} ${darkSubFormClass}`}>
      <FormHeader
        icon={location}
        title={"Delivery Address"}
        subTtile={"Where should we deliver your order?"}
        arbTitle={"عنوان التسليم"}
        arbSubTtile={"أين يجب توصيل طلبك؟"}
      />

      <div className={formInputsContainerClass}>
        <ChoicePicker
          label={isAr ? "المدينة" : "City"}
          placeholder={isAr ? "اختر المدينة" : "Choose City"}
          value={city}
          options={cities}
          error={error(CITY)}
          // A new city clears the previously chosen area.
          onSelect={(value) => setField(CITY, value, { [AREA]: "" })}
        />

        <ChoicePicker
          label={isAr ? "المنطقة" : "Area"}
          placeholder={city ? (isAr ? "اختر المنطقة" : "Choose Area") : isAr ? "اختر المدينة أولاً" : "Choose a city first"}
          value={orderFormFields[AREA].data}
          options={areas}
          error={error(AREA)}
          disabled={!city}
          onSelect={(value) => setField(AREA, value)}
        />

        <div className={FormInputContainerClass}>
          <div>
            <label>{isAr ? "تفاصيل الموقع" : "Address Details"}</label>
            <span>*</span>
          </div>
          <div className={beforeLastClass}>
            <span>
              <i className="fa-solid fa-location-dot"></i>
            </span>
            <input
              value={orderFormFields[ADDRESS].data}
              onChange={(e) => setField(ADDRESS, e.target.value)}
              placeholder={isAr ? "الشارع، رقم البناية، الطابق…" : "Street, building number, floor…"}
            />
          </div>
          <span className={errorMsgClass}>{error(ADDRESS)}</span>
        </div>
      </div>
    </div>
  );
}

export default ShippingFormForm2;
