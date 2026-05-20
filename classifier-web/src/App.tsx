import {
  useMutation,
  useQuery,
  useQueryClient
} from "@tanstack/react-query";
import { useState } from "react";
import { apiClient, type ClassListItem } from "./api/client";

type ClassStatus = ClassListItem["status"];
type GoldenPolarity = "positive" | "negative";

const classesQueryKey = ["classes"];

function App() {
  const queryClient = useQueryClient();
  const [openedClassId, setOpenedClassId] = useState<number | null>(null);

  const classesQuery = useQuery({
    queryKey: classesQueryKey,
    queryFn: loadClasses
  });

  const createClassMutation = useMutation({
    mutationFn: createClass,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: classesQueryKey });
    }
  });

  const statusMutation = useMutation({
    mutationFn: updateClassStatus,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: classesQueryKey });
    }
  });

  const goldensMutation = useMutation({
    mutationFn: updateClassGoldens,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: classesQueryKey });
    }
  });

  const busy =
    classesQuery.isFetching ||
    createClassMutation.isPending ||
    statusMutation.isPending ||
    goldensMutation.isPending;

  return (
    <main className="page">
      <section className="panel">
        <div className="pageHeader">
          <div>
            <p className="eyebrow">Classifier Orchestrator</p>
            <h1>Классы</h1>
          </div>
          <div className="loadState">{busy ? "Обновление" : "Готово"}</div>
        </div>

        <CreateClassForm
          isPending={createClassMutation.isPending}
          onCreate={(payload) => createClassMutation.mutate(payload)}
        />

        {classesQuery.isLoading && <p>Загружаем список классов...</p>}

        {classesQuery.isError && (
          <p className="error">
            Не удалось получить классы. Проверь, что orchestrator запущен на
            порту 8080.
          </p>
        )}

        {createClassMutation.isError && (
          <p className="error">Не удалось создать класс.</p>
        )}

        {statusMutation.isError && (
          <p className="error">Не удалось изменить статус класса.</p>
        )}

        {goldensMutation.isError && (
          <p className="error">Не удалось сохранить изменения голденов.</p>
        )}

        {classesQuery.isSuccess && (
          <ClassList
            classes={classesQuery.data}
            openedClassId={openedClassId}
            onToggleOpen={(classId) =>
              setOpenedClassId((current) =>
                current === classId ? null : classId
              )
            }
            onStatusChange={(classItem) =>
              statusMutation.mutate({
                classId: classItem.id,
                status: nextStatus(classItem.status)
              })
            }
            onGoldenAdd={(classItem, polarity, text) =>
              goldensMutation.mutate(
                buildGoldenUpdate(classItem, {
                  action: "add",
                  polarity,
                  text
                })
              )
            }
            onGoldenDelete={(classItem, polarity, goldenId) =>
              goldensMutation.mutate(
                buildGoldenUpdate(classItem, {
                  action: "delete",
                  polarity,
                  goldenId
                })
              )
            }
            isMutatingStatus={statusMutation.isPending}
            isMutatingGoldens={goldensMutation.isPending}
          />
        )}
      </section>
    </main>
  );
}

async function loadClasses() {
  const { data, error } = await apiClient.GET("/api/v1/classes");

  if (error) {
    throw new Error("Failed to load classes");
  }

  return data ?? [];
}

async function createClass(input: {
  name: string;
  positiveGolden: string;
  negativeGolden: string;
}) {
  const negativeGoldens = input.negativeGolden.trim()
    ? [{ text: input.negativeGolden.trim() }]
    : [];

  const { error } = await apiClient.POST("/api/v1/classes", {
    body: {
      name: input.name.trim(),
      positiveGoldens: [{ text: input.positiveGolden.trim() }],
      negativeGoldens
    }
  });

  if (error) {
    throw new Error("Failed to create class");
  }
}

async function updateClassStatus(input: {
  classId: number;
  status: ClassStatus;
}) {
  const { error } = await apiClient.PATCH("/api/v1/classes/{classId}/status", {
    params: {
      path: {
        classId: input.classId
      }
    },
    body: {
      status: input.status
    }
  });

  if (error) {
    throw new Error("Failed to update class status");
  }
}

async function updateClassGoldens(input: {
  classId: number;
  name: string;
  description: string;
  positiveGoldens: string[];
  negativeGoldens: string[];
}) {
  const { error } = await apiClient.PATCH("/api/v1/classes/{classId}", {
    params: {
      path: {
        classId: input.classId
      }
    },
    body: {
      name: input.name,
      description: input.description,
      positiveGoldens: input.positiveGoldens.map((text) => ({ text })),
      negativeGoldens: input.negativeGoldens.map((text) => ({ text }))
    }
  });

  if (error) {
    throw new Error("Failed to update class");
  }
}

function ClassList(props: {
  classes: ClassListItem[];
  openedClassId: number | null;
  onToggleOpen: (classId: number) => void;
  onStatusChange: (classItem: ClassListItem) => void;
  onGoldenAdd: (
    classItem: ClassListItem,
    polarity: GoldenPolarity,
    text: string
  ) => void;
  onGoldenDelete: (
    classItem: ClassListItem,
    polarity: GoldenPolarity,
    goldenId: number
  ) => void;
  isMutatingStatus: boolean;
  isMutatingGoldens: boolean;
}) {
  if (props.classes.length === 0) {
    return <p className="emptyState">Классы пока не созданы.</p>;
  }

  return (
    <div className="classList">
      {props.classes.map((item) => (
        <ClassCard
          classItem={item}
          isOpen={props.openedClassId === item.id}
          key={item.id}
          onToggleOpen={() => props.onToggleOpen(item.id)}
          onStatusChange={() => props.onStatusChange(item)}
          onGoldenAdd={(polarity, text) =>
            props.onGoldenAdd(item, polarity, text)
          }
          onGoldenDelete={(polarity, goldenId) =>
            props.onGoldenDelete(item, polarity, goldenId)
          }
          isMutatingStatus={props.isMutatingStatus}
          isMutatingGoldens={props.isMutatingGoldens}
        />
      ))}
    </div>
  );
}

function ClassCard(props: {
  classItem: ClassListItem;
  isOpen: boolean;
  onToggleOpen: () => void;
  onStatusChange: () => void;
  onGoldenAdd: (polarity: GoldenPolarity, text: string) => void;
  onGoldenDelete: (polarity: GoldenPolarity, goldenId: number) => void;
  isMutatingStatus: boolean;
  isMutatingGoldens: boolean;
}) {
  const item = props.classItem;

  return (
    <article
      className={`classCard ${item.status === "ACTIVE" ? "activeClass" : ""}`}
    >
      <div
        className="classRow"
        onClick={props.onToggleOpen}
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            props.onToggleOpen();
          }
        }}
        role="button"
        tabIndex={0}
      >
        <div className="classMain">
          <h2>{item.name}</h2>
          <p>{item.description || "Без описания"}</p>
        </div>

        <div className="classCounters">
          <span>+{item.positiveGoldens.length}</span>
          <span>-{item.negativeGoldens.length}</span>
        </div>

        <StatusBadge status={item.status} />

        <div className="rowActions">
          <button
            className="secondaryButton"
            disabled={props.isMutatingStatus}
            onClick={(event) => {
              event.stopPropagation();
              props.onStatusChange();
            }}
            type="button"
          >
            {item.status === "ACTIVE" ? "В DRAFT" : "В ACTIVE"}
          </button>
          <button
            className="dangerIconButton"
            disabled
            onClick={(event) => event.stopPropagation()}
            title="Удаление класса появится после добавления backend-ручки"
            type="button"
          >
            ×
          </button>
        </div>
      </div>

      {props.isOpen && (
        <ClassDetails
          classItem={item}
          isMutatingGoldens={props.isMutatingGoldens}
          onGoldenAdd={props.onGoldenAdd}
          onGoldenDelete={props.onGoldenDelete}
        />
      )}
    </article>
  );
}

function ClassDetails(props: {
  classItem: ClassListItem;
  isMutatingGoldens: boolean;
  onGoldenAdd: (polarity: GoldenPolarity, text: string) => void;
  onGoldenDelete: (polarity: GoldenPolarity, goldenId: number) => void;
}) {
  const item = props.classItem;

  return (
    <div className="classDetails">
      <dl className="detailsGrid">
        <div>
          <dt>ID</dt>
          <dd>{item.id}</dd>
        </div>
        <div>
          <dt>Создан</dt>
          <dd>{formatDate(item.createdAt)}</dd>
        </div>
        <div>
          <dt>Обновлен</dt>
          <dd>{formatDate(item.updatedAt)}</dd>
        </div>
      </dl>

      <div className="goldenColumns">
        <GoldenSection
          title="Положительные голдены"
          polarity="positive"
          goldens={item.positiveGoldens}
          isDeleteDisabled={(goldenId) =>
            props.isMutatingGoldens ||
            item.positiveGoldens.length <= 1 ||
            !item.positiveGoldens.some((golden) => golden.id === goldenId)
          }
          isMutating={props.isMutatingGoldens}
          onAdd={(text) => props.onGoldenAdd("positive", text)}
          onDelete={(goldenId) => props.onGoldenDelete("positive", goldenId)}
        />
        <GoldenSection
          title="Отрицательные голдены"
          polarity="negative"
          goldens={item.negativeGoldens}
          isDeleteDisabled={() => props.isMutatingGoldens}
          isMutating={props.isMutatingGoldens}
          onAdd={(text) => props.onGoldenAdd("negative", text)}
          onDelete={(goldenId) => props.onGoldenDelete("negative", goldenId)}
        />
      </div>
    </div>
  );
}

function GoldenSection(props: {
  title: string;
  polarity: GoldenPolarity;
  goldens: ClassListItem["positiveGoldens"];
  isMutating: boolean;
  isDeleteDisabled: (goldenId: number) => boolean;
  onAdd: (text: string) => void;
  onDelete: (goldenId: number) => void;
}) {
  const [text, setText] = useState("");
  const trimmedText = text.trim();

  return (
    <section className="goldenSection">
      <div className="sectionHeader">
        <h3>{props.title}</h3>
        <span className={`polarityMark ${props.polarity}`}>
          {props.goldens.length}
        </span>
      </div>

      {props.goldens.length === 0 ? (
        <p className="muted">Список пуст.</p>
      ) : (
        <ul className="goldenList">
          {props.goldens.map((golden) => (
            <li key={golden.id}>
              <span>{golden.text}</span>
              <button
                className="smallDangerButton"
                disabled={props.isDeleteDisabled(golden.id)}
                onClick={() => props.onDelete(golden.id)}
                type="button"
              >
                Удалить
              </button>
            </li>
          ))}
        </ul>
      )}

      <form
        className="goldenForm"
        onSubmit={(event) => {
          event.preventDefault();

          if (!trimmedText) {
            return;
          }

          props.onAdd(trimmedText);
          setText("");
        }}
      >
        <input
          disabled={props.isMutating}
          onChange={(event) => setText(event.target.value)}
          placeholder="Текст нового голдена"
          type="text"
          value={text}
        />
        <button
          className="primaryButton"
          disabled={!trimmedText || props.isMutating}
          type="submit"
        >
          Добавить
        </button>
      </form>
    </section>
  );
}

function CreateClassForm(props: {
  isPending: boolean;
  onCreate: (input: {
    name: string;
    positiveGolden: string;
    negativeGolden: string;
  }) => void;
}) {
  const [name, setName] = useState("");
  const [positiveGolden, setPositiveGolden] = useState("");
  const [negativeGolden, setNegativeGolden] = useState("");
  const canSubmit = name.trim() && positiveGolden.trim() && !props.isPending;

  return (
    <form
      className="createClassForm"
      onSubmit={(event) => {
        event.preventDefault();

        if (!canSubmit) {
          return;
        }

        props.onCreate({
          name,
          positiveGolden,
          negativeGolden
        });
        setName("");
        setPositiveGolden("");
        setNegativeGolden("");
      }}
    >
      <input
        disabled={props.isPending}
        onChange={(event) => setName(event.target.value)}
        placeholder="Название класса"
        type="text"
        value={name}
      />
      <input
        disabled={props.isPending}
        onChange={(event) => setPositiveGolden(event.target.value)}
        placeholder="Положительный голден"
        type="text"
        value={positiveGolden}
      />
      <input
        disabled={props.isPending}
        onChange={(event) => setNegativeGolden(event.target.value)}
        placeholder="Отрицательный голден"
        type="text"
        value={negativeGolden}
      />
      <button className="primaryButton" disabled={!canSubmit} type="submit">
        Создать класс
      </button>
    </form>
  );
}

function StatusBadge({ status }: { status: ClassStatus }) {
  return <span className={`statusBadge ${status.toLowerCase()}`}>{status}</span>;
}

function nextStatus(status: ClassStatus): ClassStatus {
  return status === "ACTIVE" ? "DRAFT" : "ACTIVE";
}

function buildGoldenUpdate(
  classItem: ClassListItem,
  change:
    | {
        action: "add";
        polarity: GoldenPolarity;
        text: string;
      }
    | {
        action: "delete";
        polarity: GoldenPolarity;
        goldenId: number;
      }
) {
  const positiveGoldens = classItem.positiveGoldens.map((golden) => golden.text);
  const negativeGoldens = classItem.negativeGoldens.map((golden) => golden.text);

  if (change.action === "add" && change.text.trim()) {
    if (change.polarity === "positive") {
      positiveGoldens.push(change.text.trim());
    } else {
      negativeGoldens.push(change.text.trim());
    }
  }

  if (change.action === "delete") {
    const source =
      change.polarity === "positive" ? positiveGoldens : negativeGoldens;
    const currentGoldens =
      change.polarity === "positive"
        ? classItem.positiveGoldens
        : classItem.negativeGoldens;
    const index = currentGoldens.findIndex(
      (golden) => golden.id === change.goldenId
    );

    if (index >= 0) {
      source.splice(index, 1);
    }
  }

  return {
    classId: classItem.id,
    name: classItem.name,
    description: classItem.description,
    positiveGoldens,
    negativeGoldens
  };
}

function formatDate(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}

export default App;
