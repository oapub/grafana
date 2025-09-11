import { css } from '@emotion/css';

import { GrafanaTheme2 } from '@grafana/data';
import { t } from '@grafana/i18n';
import { sceneGraph } from '@grafana/scenes';
import { Box, Dropdown, Menu, ToolbarButton, useStyles2 } from '@grafana/ui';

import { DashboardLinkRenderer } from './DashboardLinkRenderer';
import { DashboardScene } from './DashboardScene';
import { VariableValueSelectWrapper } from './VariableControls';

export const DASHBOARD_CONTROLS_MENU_ARIA_LABEL = 'Dashboard controls menu';
export const DASHBOARD_CONTROLS_MENU_TITLE = 'Dashboard controls';

export function DashboardControlsMenu({ dashboard }: { dashboard: DashboardScene }) {
  const styles = useStyles2(getStyles);
  const { links, uid } = dashboard.useState();
  // Dashboard links as dropdowns aren't supported here.
  // Reason: nesting <Dropdown> components causes issues since the inner dropdown is rendered in a portal,
  // so clicking it closes the parent dropdown (the parent sees it as an overlay click, and the event cannot easily be intercepted,
  // as it is in different HTML subtree).
  const filteredLinks = links.filter((link) => link.placement === 'inControlsMenu' && !link.asDropdown);
  const variables = sceneGraph
    .getVariables(dashboard)!
    .useState()
    .variables.filter((v) => v.state.showInControlsMenu === true);

  if ((variables.length === 0 && filteredLinks.length === 0) || !uid) {
    return null;
  }

  return (
    <Dropdown
      overlay={
        <Box
          backgroundColor="elevated"
          borderRadius="default"
          borderColor="medium"
          borderStyle="solid"
          boxShadow="z3"
          display="inline-block"
          paddingX={1}
          paddingY={2}
          marginRight={1}
          role="menu"
          tabIndex={-1}
          width={40}
          onClick={(e) => {
            // Normally, clicking the overlay closes the dropdown.
            // We stop event propagation here to keep it open while users interact with variable controls.
            e.stopPropagation();
          }}
        >
          {/* Variables */}
          {variables.map((variable) => (
            <div className={styles.variableItem} key={variable.state.key}>
              <VariableValueSelectWrapper variable={variable} layout="vertical" />
            </div>
          ))}

          {variables.length > 0 && filteredLinks.length > 0 && (
            <div className={styles.divider}>
              <Menu.Divider />
            </div>
          )}

          {/* Links */}
          {filteredLinks.map((link, index) => (
            <div className={styles.linkItem} key={`${link.title}-$${index}`}>
              <DashboardLinkRenderer link={link} dashboardUID={uid} buttonFill="text" />
            </div>
          ))}
        </Box>
      }
    >
      <ToolbarButton
        aria-label={t('dashboard.controls.menu.aria-label', DASHBOARD_CONTROLS_MENU_ARIA_LABEL)}
        title={t('dashboard.controls.menu.title', DASHBOARD_CONTROLS_MENU_TITLE)}
        icon="ellipsis-v"
        iconSize="md"
        narrow
      />
    </Dropdown>
  );
}

const getStyles = (theme: GrafanaTheme2) => ({
  variableItem: css({
    '&:not(:first-child)': {
      marginTop: theme.spacing(2),
    },
    padding: theme.spacing(0, 0.5),
  }),
  linkItem: css({
    // '&:not(:first-child)': {
    //   marginTop: theme.spacing(1),
    // },
  }),
  divider: css({
    margin: theme.spacing(2, 0),
    padding: theme.spacing(0, 0.5),
  }),
});
