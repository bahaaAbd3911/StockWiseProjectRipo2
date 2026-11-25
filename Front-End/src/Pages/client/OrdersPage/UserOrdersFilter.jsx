import { UserOrdersFilterContainerClass, filterContainer, activeFilterContainer } from "./ClientOrdersPage.module.css";
import { useLang } from "../../../hooks/useLang";
import { ORDER_GROUPS, inGroup } from "./orderGroups";

function UserOrdersFilter({ orders, selected, onSelect }) {
  const { isAr } = useLang();

  return (
    <div className={UserOrdersFilterContainerClass} role="tablist">
      {ORDER_GROUPS.map((group) => (
        <button
          type="button"
          role="tab"
          aria-selected={selected === group.key}
          className={`${filterContainer} ${selected === group.key ? activeFilterContainer : ""}`}
          key={group.key}
          onClick={() => onSelect(group.key)}
        >
          <i className={group.icon}></i>
          <span>{isAr ? group.arbName : group.name}</span>
          <span>{orders.filter((o) => inGroup(o, group)).length}</span>
        </button>
      ))}
    </div>
  );
}

export default UserOrdersFilter;
