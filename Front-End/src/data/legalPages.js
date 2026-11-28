// Texts of the Privacy Policy and Terms pages (/client/privacy, /client/terms).
// {store} is replaced by the store's name.
//
// They describe what this platform really does with data (see docs/07, 08,
// 13). They are a starting point, NOT legal advice: have them reviewed for
// the country the store sells in before relying on them.

export const LEGAL_UPDATED = "2026-11-28";

export const PRIVACY = {
  title: ["Privacy Policy", "سياسة الخصوصية"],
  intro: [
    "This page explains what {store} collects about you, why, and what you can do about it.",
    "توضح هذه الصفحة ما يجمعه {store} من معلومات عنك ولماذا، وما يمكنك فعله بشأنها.",
  ],
  sections: [
    {
      title: ["What we collect", "ما نجمعه"],
      items: [
        [
          "Your account: name, email address, phone number and, if you add one, a profile picture.",
          "حسابك: الاسم والبريد الإلكتروني ورقم الهاتف، وصورة شخصية إن أضفتها.",
        ],
        [
          "Delivery addresses you save, and the address copied onto each order.",
          "عناوين التوصيل التي تحفظها، والعنوان المنسوخ على كل طلب.",
        ],
        [
          "Your orders, cart, favorites, reviews and waiting-list requests.",
          "طلباتك وسلة التسوق والمفضلة والتقييمات وطلبات قائمة الانتظار.",
        ],
        [
          "If the store lets you pay later: your credit limit, what you owe, and your payments.",
          "إذا سمح لك المتجر بالدفع لاحقاً: حد الائتمان والمبلغ المستحق ودفعاتك.",
        ],
        [
          "Technical data needed to keep the site safe, such as your IP address when you ask for a password reset.",
          "بيانات تقنية لازمة لحماية الموقع، مثل عنوان IP عند طلب إعادة تعيين كلمة المرور.",
        ],
      ],
    },
    {
      title: ["Why we use it", "لماذا نستخدمها"],
      items: [
        [
          "To deliver your orders and contact you about them.",
          "لتوصيل طلباتك والتواصل معك بشأنها.",
        ],
        [
          "To send the emails you need: confirming your address, order updates, payments, back-in-stock notices and balance reminders.",
          "لإرسال الرسائل اللازمة: تأكيد بريدك، تحديثات الطلب، الدفعات، إشعارات عودة المنتجات وتذكيرات الرصيد.",
        ],
        [
          "To protect accounts, for example by limiting repeated login attempts.",
          "لحماية الحسابات، مثل تقييد محاولات تسجيل الدخول المتكررة.",
        ],
      ],
    },
    {
      title: ["What we do not do", "ما لا نفعله"],
      items: [
        [
          "We do not sell your data and do not use advertising or tracking cookies.",
          "لا نبيع بياناتك ولا نستخدم ملفات تعريف الارتباط الإعلانية أو للتتبع.",
        ],
        [
          "Passwords are stored only in a scrambled (hashed) form that cannot be read back.",
          "تُخزَّن كلمات المرور بشكل مُشفَّر لا يمكن استرجاعه.",
        ],
        [
          "Photos you upload are cleaned of hidden information, such as the place where they were taken.",
          "تُزال من الصور التي ترفعها المعلومات المخفية، مثل مكان التقاطها.",
        ],
      ],
    },
    {
      title: ["Your browser", "متصفحك"],
      items: [
        [
          "Your browser keeps your login, your language and theme choice, and the store's look so pages load faster. Signing out removes the login.",
          "يحفظ متصفحك تسجيل دخولك واختيارك للغة والمظهر وشكل المتجر لتسريع التحميل. تسجيل الخروج يحذف تسجيل الدخول.",
        ],
        [
          "Fonts and icons are loaded from Google Fonts and cdnjs, which see your IP address like any website you visit.",
          "تُحمَّل الخطوط والأيقونات من Google Fonts وcdnjs، اللذين يريان عنوان IP مثل أي موقع تزوره.",
        ],
      ],
    },
    {
      title: ["Who else handles your data", "من يتعامل مع بياناتك أيضاً"],
      items: [
        [
          "The company that hosts the website, and the email provider that delivers our emails.",
          "الشركة المستضيفة للموقع، ومزود البريد الذي يوصل رسائلنا.",
        ],
        [
          "Delivery staff receive your name, phone number and delivery address for your order.",
          "يحصل موظفو التوصيل على اسمك ورقم هاتفك وعنوان التوصيل لطلبك.",
        ],
      ],
    },
    {
      title: ["How long we keep it", "مدة الاحتفاظ بها"],
      items: [
        [
          "Orders are kept as business records, including after an account is closed.",
          "تُحفظ الطلبات كسجلات تجارية، حتى بعد إغلاق الحساب.",
        ],
        [
          "Delivered emails are deleted from our system after 30 days.",
          "تُحذف الرسائل المرسلة من نظامنا بعد 30 يوماً.",
        ],
        [
          "Password-reset and confirmation links expire after a short time and work only once.",
          "تنتهي صلاحية روابط إعادة التعيين والتأكيد بعد وقت قصير وتعمل مرة واحدة فقط.",
        ],
      ],
    },
    {
      title: ["Your choices", "خياراتك"],
      items: [
        [
          "You can see and change your profile and addresses at any time in your profile.",
          "يمكنك عرض ملفك الشخصي وعناوينك وتعديلها في أي وقت.",
        ],
        [
          "You can close your account from your profile. To ask for a copy or deletion of your data, contact the store.",
          "يمكنك إغلاق حسابك من ملفك الشخصي. لطلب نسخة من بياناتك أو حذفها، تواصل مع المتجر.",
        ],
      ],
    },
  ],
};

export const TERMS = {
  title: ["Terms of Use", "شروط الاستخدام"],
  intro: [
    "By creating an account or placing an order with {store}, you agree to these terms.",
    "بإنشاء حساب أو تقديم طلب لدى {store}، فإنك توافق على هذه الشروط.",
  ],
  sections: [
    {
      title: ["Your account", "حسابك"],
      items: [
        [
          "Give correct details and keep your password private. You are responsible for orders placed from your account.",
          "قدّم بيانات صحيحة وحافظ على سرية كلمة المرور. أنت مسؤول عن الطلبات المقدمة من حسابك.",
        ],
        [
          "You must confirm your email address before placing an order.",
          "يجب تأكيد بريدك الإلكتروني قبل تقديم أي طلب.",
        ],
      ],
    },
    {
      title: ["Prices and orders", "الأسعار والطلبات"],
      items: [
        [
          "Prices are shown in the store's currency and include any offer that applies at the moment you order. The total at checkout, including shipping, is what you pay.",
          "تُعرض الأسعار بعملة المتجر وتشمل أي عرض سارٍ لحظة الطلب. المجموع عند إتمام الطلب، شاملاً الشحن، هو ما تدفعه.",
        ],
        [
          "An order is confirmed when the store accepts it. The store may cancel an order, for example if a product cannot be delivered.",
          "يُؤكَّد الطلب عندما يقبله المتجر. يجوز للمتجر إلغاء الطلب، مثلاً إذا تعذّر توصيل منتج.",
        ],
        [
          "Coupons and offers have their own conditions and dates, shown when you use them.",
          "للقسائم والعروض شروط وتواريخ خاصة تظهر عند استخدامها.",
        ],
      ],
    },
    {
      title: ["Payment", "الدفع"],
      items: [
        [
          "Cash on delivery: pay when you receive the order.",
          "الدفع عند الاستلام: ادفع عند استلام الطلب.",
        ],
        [
          "Pay later (only if the store offers it to you): the order total is added to your balance, which you agree to pay back to the store.",
          "الدفع لاحقاً (فقط إذا أتاحه لك المتجر): يُضاف مبلغ الطلب إلى رصيدك، وتتعهد بسداده للمتجر.",
        ],
      ],
    },
    {
      title: ["Cancelling", "الإلغاء"],
      items: [
        [
          "You can cancel an order yourself while it is still pending. After that, contact the store.",
          "يمكنك إلغاء الطلب بنفسك ما دام قيد الانتظار. بعد ذلك تواصل مع المتجر.",
        ],
      ],
    },
    {
      title: ["Reviews", "التقييمات"],
      items: [
        [
          "Reviews must be honest and polite. The store may remove reviews that are not.",
          "يجب أن تكون التقييمات صادقة ولائقة، ويحق للمتجر حذف ما يخالف ذلك.",
        ],
      ],
    },
    {
      title: ["Changes", "التعديلات"],
      items: [
        [
          "These terms may be updated; the date below shows the latest version.",
          "قد تُحدَّث هذه الشروط، ويوضح التاريخ أدناه آخر نسخة.",
        ],
      ],
    },
  ],
};
