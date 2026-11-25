import { useCallback, useEffect, useState } from "react";

import * as domainsApi from "../../../api/domains";
import { Button, Card, ConfirmDialog, EmptyState, FormField, PageHeader, StatusBadge } from "../../../components/admin/ui";
import { useApiAction } from "../../../hooks/useApiAction";
import { useLang } from "../../../hooks/useLang";
import styles from "./Domains.module.css";

// A store's own domain names (e.g. shop.example.com):
//   1. add it  2. create the two DNS records shown  3. verify
// Verified domains show this store, get HTTPS automatically, and one of them
// can be the main address used in emails.
export default function Domains() {
  const { tr } = useLang();
  const { run, busy } = useApiAction();
  const [domains, setDomains] = useState([]);
  const [input, setInput] = useState("");
  const [removing, setRemoving] = useState(null);

  const load = useCallback(() => domainsApi.fetchDomains(), []);
  const reload = async () => setDomains(await load());

  useEffect(() => {
    let current = true;
    load().then((list) => current && setDomains(list), () => {});
    return () => {
      current = false;
    };
  }, [load]);

  const add = (e) => {
    e.preventDefault();
    if (!input.trim()) return;
    run(async () => {
      await domainsApi.addDomain(input.trim());
      setInput("");
      await reload();
    }, ["Domain added: now create the DNS records", "تمت إضافة النطاق: أنشئ سجلات DNS الآن"]);
  };

  return (
    <>
      <PageHeader
        icon="fa-solid fa-globe"
        title={tr("Domains", "النطاقات")}
        subtitle={tr("Open your store on your own domain name.", "افتح متجرك على اسم نطاقك الخاص.")}
      />

      <Card className={styles.addCard}>
        <form onSubmit={add} className={styles.addForm} noValidate>
          <FormField label={tr("Domain", "النطاق")} name="domain" value={input} placeholder="shop.example.com"
            onChange={(e) => setInput(e.target.value)} />
          <Button type="submit" icon="fa-solid fa-plus" disabled={busy}>
            {tr("Add domain", "إضافة نطاق")}
          </Button>
        </form>
      </Card>

      {domains.length === 0 ? (
        <EmptyState icon="fa-solid fa-globe" text={tr("No domains yet. Your store is always available at its platform address.", "لا توجد نطاقات بعد. متجرك متاح دائماً على عنوانه في المنصة.")} />
      ) : (
        domains.map((d) => (
          <Card key={d.id} className={styles.domainCard}>
            <div className={styles.domainHead}>
              <strong>{d.domain}</strong>
              {d.verified ? (
                <StatusBadge tone="green">{tr("Verified", "تم التحقق")}</StatusBadge>
              ) : (
                <StatusBadge tone="orange">{tr("Waiting for DNS", "بانتظار DNS")}</StatusBadge>
              )}
              {d.is_primary && <StatusBadge tone="blue">{tr("Main address", "العنوان الرئيسي")}</StatusBadge>}
              <span className={styles.spacer} />
              {!d.verified && (
                <Button icon="fa-solid fa-rotate" disabled={busy}
                  onClick={() => run(async () => { await domainsApi.verifyDomain(d.id); await reload(); }, ["Domain verified", "تم التحقق من النطاق"])}>
                  {tr("Verify", "تحقق")}
                </Button>
              )}
              {d.verified && !d.is_primary && (
                <Button variant="outline" disabled={busy}
                  onClick={() => run(async () => { await domainsApi.makePrimaryDomain(d.id); await reload(); }, ["Main address changed", "تم تغيير العنوان الرئيسي"])}>
                  {tr("Make main address", "اجعله العنوان الرئيسي")}
                </Button>
              )}
              <Button variant="ghost" icon="fa-solid fa-trash" onClick={() => setRemoving(d)}>
                {tr("Remove", "إزالة")}
              </Button>
            </div>

            {!d.verified && (
              <div className={styles.dns}>
                <p>{tr("At your domain provider, create these two records, then press Verify (DNS can take a few hours):", "لدى مزود النطاق أنشئ هذين السجلين ثم اضغط تحقق (قد يستغرق DNS بضع ساعات):")}</p>
                <table>
                  <thead>
                    <tr>
                      <th>{tr("Type", "النوع")}</th>
                      <th>{tr("Name", "الاسم")}</th>
                      <th>{tr("Value", "القيمة")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>TXT</td>
                      <td><code>{d.dns.txt_name}</code></td>
                      <td><code>{d.dns.txt_value}</code></td>
                    </tr>
                    <tr>
                      <td>CNAME</td>
                      <td><code>{d.domain}</code></td>
                      <td><code>{d.dns.cname_target}</code></td>
                    </tr>
                  </tbody>
                </table>
                <p className={styles.note}>
                  {tr("For a root domain (example.com without www), use an ALIAS/ANAME record or the server's IP address instead of CNAME.",
                    "للنطاق الجذري (بدون www) استخدم سجل ALIAS/ANAME أو عنوان IP الخادم بدل CNAME.")}
                </p>
              </div>
            )}
          </Card>
        ))
      )}

      {removing && (
        <ConfirmDialog
          title={tr("Remove domain", "إزالة النطاق")}
          message={tr(`${removing.domain} will stop showing your store.`, `سيتوقف ${removing.domain} عن عرض متجرك.`)}
          confirmLabel={tr("Remove", "إزالة")}
          onConfirm={() => run(async () => { await domainsApi.removeDomain(removing.id); setRemoving(null); await reload(); }, ["Domain removed", "تمت إزالة النطاق"])}
          onCancel={() => setRemoving(null)}
        />
      )}
    </>
  );
}
