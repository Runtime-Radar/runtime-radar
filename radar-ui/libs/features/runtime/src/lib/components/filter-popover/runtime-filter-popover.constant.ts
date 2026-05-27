import { RuntimeEventFilterKey } from '../../interfaces/runtime-filter.interface';

export const RUNTIME_FILTER_POPOVER_VISIBLE_CONTROL_KEYS: RuntimeEventFilterKey[] = [
    RuntimeEventFilterKey.NAMESPACE,
    RuntimeEventFilterKey.POD
];

export const RUNTIME_FILTER_POPOVER_CONTROL_LABELS = new Map<RuntimeEventFilterKey, string>([
    [RuntimeEventFilterKey.TYPE, 'Runtime.EventsPage.Filter.FilterPopover.Label.Type'],
    [RuntimeEventFilterKey.ARGUMENT, 'Runtime.EventsPage.Filter.FilterPopover.Label.Argument'],
    [RuntimeEventFilterKey.BINARY, 'Runtime.EventsPage.Filter.FilterPopover.Label.Binary'],
    [RuntimeEventFilterKey.CONTAINER, 'Runtime.EventsPage.Filter.FilterPopover.Label.Container'],
    [RuntimeEventFilterKey.FUNCTION, 'Runtime.EventsPage.Filter.FilterPopover.Label.Function'],
    [RuntimeEventFilterKey.IMAGE, 'Runtime.EventsPage.Filter.FilterPopover.Label.Image'],
    [RuntimeEventFilterKey.NAMESPACE, 'Runtime.EventsPage.Filter.FilterPopover.Label.Namespace'],
    [RuntimeEventFilterKey.POD, 'Runtime.EventsPage.Filter.FilterPopover.Label.Pod'],
    [RuntimeEventFilterKey.PERIOD, 'Runtime.EventsPage.Filter.FilterPopover.Label.Period']
]);

export const RUNTIME_FILTER_POPOVER_CONTROL_PLACEHOLDERS = new Map<RuntimeEventFilterKey, string>([
    [RuntimeEventFilterKey.TYPE, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Type'],
    [RuntimeEventFilterKey.ARGUMENT, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Argument'],
    [RuntimeEventFilterKey.BINARY, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Binary'],
    [RuntimeEventFilterKey.CONTAINER, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Container'],
    [RuntimeEventFilterKey.FUNCTION, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Function'],
    [RuntimeEventFilterKey.IMAGE, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Image'],
    [RuntimeEventFilterKey.NAMESPACE, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Namespace'],
    [RuntimeEventFilterKey.POD, 'Runtime.EventsPage.Filter.FilterPopover.Placeholder.Pod']
]);
