import { getOpenAPIClient } from './openapi-client';
import {
  ApiResponseError,
  unwrapData,
  unwrapPaged,
  unwrapVoid,
  type ListOptions,
} from './attempts';
import type { components, paths } from './openapi-schema';

export type AssessmentListItem = components['schemas']['AssessmentListItem'];
export type PagedAssessments =
  paths['/assessments']['get']['responses'][200]['content']['application/json'];

export type AssessmentDetail = components['schemas']['AssessmentDetail']['data'];
export type Section = components['schemas']['Section']['data'];
export type Item = components['schemas']['Item']['data'];
export type Target = components['schemas']['Target']['data'];
export type CreateAssessmentRequest =
  components['schemas']['CreateAssessmentRequest'];
export type UpdateAssessmentRequest =
  components['schemas']['UpdateAssessmentRequest'];
export type CreateSectionRequest =
  components['schemas']['CreateSectionRequest'];
export type CreateItemRequest = components['schemas']['CreateItemRequest'];
export type CreateTargetRequest =
  components['schemas']['CreateTargetRequest'];
export type ValidationResult =
  components['schemas']['ValidationResult']['data'];
export type PublishResult = components['schemas']['PublishResult']['data'];
export type QuestionPickerItem =
  components['schemas']['QuestionPickerItem'];
export type PagedQuestions =
  paths['/questions']['get']['responses'][200]['content']['application/json'];
export type PublicationSummary =
  components['schemas']['PublicationSummary'];
export type UpdateSectionRequest =
  components['schemas']['UpdateSectionRequest'];
export type UpdateItemRequest =
  components['schemas']['UpdateItemRequest'];
export type ReorderSectionsRequest =
  components['schemas']['ReorderSectionsRequest'];
export type ReorderItemsRequest =
  components['schemas']['ReorderItemsRequest'];
export type AssessmentPreview =
  components['schemas']['AssessmentPreview']['data'];
export type PreviewSection =
  components['schemas']['PreviewSection']['data'];
export type PreviewItem = components['schemas']['PreviewItem']['data'];

function cleanListQuery(opts: ListOptions) {
  return {
    q: opts.q,
    limit: opts.limit,
    offset: opts.offset,
    cursor: opts.cursor,
    count: opts.count,
  };
}

function cleanQuestionQuery(
  opts: { q?: string; bank_id?: string; limit?: number; offset?: number }
) {
  return {
    q: opts.q,
    bank_id: opts.bank_id,
    limit: opts.limit,
    offset: opts.offset,
  };
}

export async function listAssessments(
  opts: ListOptions = {}
): Promise<PagedAssessments> {
  const client = await getOpenAPIClient();
  return unwrapPaged<AssessmentListItem>(
    await client.GET('/assessments', { params: { query: cleanListQuery(opts) } })
  ) as PagedAssessments;
}

export async function createAssessment(
  classSectionId: string,
  req: CreateAssessmentRequest
): Promise<AssessmentDetail> {
  const client = await getOpenAPIClient();
  return unwrapData<AssessmentDetail>(
    await client.POST('/classes/{class_id}/assessments', {
      params: { path: { class_id: classSectionId } },
      body: req,
    })
  );
}

export async function getAssessment(id: string): Promise<AssessmentDetail> {
  const client = await getOpenAPIClient();
  return unwrapData<AssessmentDetail>(
    await client.GET('/assessments/{id}', { params: { path: { id } } })
  );
}

export async function updateAssessment(
  id: string,
  req: UpdateAssessmentRequest
): Promise<AssessmentDetail> {
  const client = await getOpenAPIClient();
  return unwrapData<AssessmentDetail>(
    await client.PATCH('/assessments/{id}', {
      params: { path: { id } },
      body: req,
    })
  );
}

export async function createSection(
  assessmentId: string,
  req: CreateSectionRequest
): Promise<Section> {
  const client = await getOpenAPIClient();
  return unwrapData<Section>(
    await client.POST('/assessments/{id}/sections', {
      params: { path: { id: assessmentId } },
      body: req,
    })
  );
}

export async function createItem(
  sectionId: string,
  req: CreateItemRequest
): Promise<Item> {
  const client = await getOpenAPIClient();
  return unwrapData<Item>(
    await client.POST('/assessment-sections/{section_id}/items', {
      params: { path: { section_id: sectionId } },
      body: req,
    })
  );
}

export async function createTarget(
  assessmentId: string,
  req: CreateTargetRequest
): Promise<Target> {
  const client = await getOpenAPIClient();
  return unwrapData<Target>(
    await client.POST('/assessments/{id}/targets', {
      params: { path: { id: assessmentId } },
      body: req,
    })
  );
}

export async function validateAssessment(id: string): Promise<ValidationResult> {
  const client = await getOpenAPIClient();
  return unwrapData<ValidationResult>(
    await client.POST('/assessments/{id}/validate', {
      params: { path: { id } },
    })
  );
}

export async function publishAssessment(id: string): Promise<PublishResult> {
  const client = await getOpenAPIClient();
  return unwrapData<PublishResult>(
    await client.POST('/assessments/{id}/publish', {
      params: { path: { id } },
    })
  );
}

export async function listQuestions(
  opts: { q?: string; bank_id?: string; limit?: number; offset?: number } = {}
): Promise<PagedQuestions> {
  const client = await getOpenAPIClient();
  return unwrapPaged<QuestionPickerItem>(
    await client.GET('/questions', {
      params: { query: cleanQuestionQuery(opts) },
    })
  ) as PagedQuestions;
}

export async function updateSection(
  sectionId: string,
  req: UpdateSectionRequest
): Promise<Section> {
  const client = await getOpenAPIClient();
  return unwrapData<Section>(
    await client.PATCH('/assessment-sections/{section_id}', {
      params: { path: { section_id: sectionId } },
      body: req,
    })
  );
}

export async function deleteSection(sectionId: string): Promise<void> {
  const client = await getOpenAPIClient();
  unwrapVoid(
    await client.DELETE('/assessment-sections/{section_id}', {
      params: { path: { section_id: sectionId } },
    })
  );
}

export async function reorderSections(
  assessmentId: string,
  req: ReorderSectionsRequest
): Promise<void> {
  const client = await getOpenAPIClient();
  unwrapVoid(
    await client.POST('/assessments/{id}/sections/reorder', {
      params: { path: { id: assessmentId } },
      body: req,
    })
  );
}

export async function updateItem(
  itemId: string,
  req: UpdateItemRequest
): Promise<Item> {
  const client = await getOpenAPIClient();
  return unwrapData<Item>(
    await client.PATCH('/assessment-items/{item_id}', {
      params: { path: { item_id: itemId } },
      body: req,
    })
  );
}

export async function deleteItem(itemId: string): Promise<void> {
  const client = await getOpenAPIClient();
  unwrapVoid(
    await client.DELETE('/assessment-items/{item_id}', {
      params: { path: { item_id: itemId } },
    })
  );
}

export async function reorderItems(
  sectionId: string,
  req: ReorderItemsRequest
): Promise<void> {
  const client = await getOpenAPIClient();
  unwrapVoid(
    await client.POST('/assessment-sections/{section_id}/items/reorder', {
      params: { path: { section_id: sectionId } },
      body: req,
    })
  );
}

export async function deleteTarget(
  assessmentId: string,
  targetId: string
): Promise<void> {
  const client = await getOpenAPIClient();
  unwrapVoid(
    await client.DELETE('/assessments/{id}/targets/{target_id}', {
      params: { path: { id: assessmentId, target_id: targetId } },
    })
  );
}

export async function listPublications(
  assessmentId: string
): Promise<PublicationSummary[]> {
  const client = await getOpenAPIClient();
  return unwrapData<PublicationSummary[]>(
    await client.GET('/assessments/{id}/publications', {
      params: { path: { id: assessmentId } },
    })
  );
}

export async function previewAssessment(id: string): Promise<AssessmentPreview> {
  const client = await getOpenAPIClient();
  return unwrapData<AssessmentPreview>(
    await client.GET('/assessments/{id}/preview', {
      params: { path: { id } },
    })
  );
}

export async function duplicateSection(
  assessmentId: string,
  sectionId: string
): Promise<Section> {
  const client = await getOpenAPIClient();
  return unwrapData<Section>(
    await client.POST('/assessments/{id}/sections/{section_id}/duplicate', {
      params: { path: { id: assessmentId, section_id: sectionId } },
    })
  );
}

export async function duplicateItem(
  sectionId: string,
  itemId: string
): Promise<Item> {
  const client = await getOpenAPIClient();
  return unwrapData<Item>(
    await client.POST(
      '/assessment-sections/{section_id}/items/{item_id}/duplicate',
      {
        params: { path: { section_id: sectionId, item_id: itemId } },
      }
    )
  );
}

// ---- Question banks ----

export type QuestionBank = components['schemas']['QuestionBank'];
export type CreateQuestionBankRequest =
  components['schemas']['CreateQuestionBankRequest'];
export type QuestionBankQuestion = components['schemas']['QuestionBankQuestion'];
export type CreateQuestionRequest = components['schemas']['CreateQuestionRequest'];
export type CreateQuestionResponse =
  components['schemas']['CreateQuestionResponse'];
export type QuestionVersion = components['schemas']['QuestionVersion'];
export type CreateQuestionVersionRequest =
  components['schemas']['CreateQuestionVersionRequest'];
export type PublishQuestionVersionResult =
  components['schemas']['PublishQuestionVersionResult'];

export async function listQuestionBanks(
  opts: { q?: string; include_archived?: boolean; limit?: number; offset?: number } = {}
): Promise<QuestionBank[]> {
  const client = await getOpenAPIClient();
  return unwrapData<QuestionBank[]>(
    await client.GET('/question-banks', {
      params: {
        query: {
          q: opts.q,
          include_archived: opts.include_archived,
          limit: opts.limit,
          offset: opts.offset,
        },
      },
    })
  );
}

export async function createQuestionBank(
  req: CreateQuestionBankRequest
): Promise<QuestionBank> {
  const client = await getOpenAPIClient();
  return unwrapData<QuestionBank>(
    await client.POST('/question-banks', { body: req })
  );
}

export async function listQuestionsInBank(
  bankId: string,
  opts: { include_archived?: boolean; limit?: number; offset?: number } = {}
): Promise<QuestionBankQuestion[]> {
  const client = await getOpenAPIClient();
  return unwrapData<QuestionBankQuestion[]>(
    await client.GET('/question-banks/{bank_id}/questions', {
      params: {
        path: { bank_id: bankId },
        query: {
          include_archived: opts.include_archived,
          limit: opts.limit,
          offset: opts.offset,
        },
      },
    })
  );
}

export async function createQuestionInBank(
  bankId: string,
  req: CreateQuestionRequest
): Promise<CreateQuestionResponse> {
  const client = await getOpenAPIClient();
  return unwrapData<CreateQuestionResponse>(
    await client.POST('/question-banks/{bank_id}/questions', {
      params: { path: { bank_id: bankId } },
      body: req,
    })
  );
}

export async function createQuestionVersion(
  bankId: string,
  questionId: string,
  req: CreateQuestionVersionRequest
): Promise<QuestionVersion> {
  const client = await getOpenAPIClient();
  return unwrapData<QuestionVersion>(
    await client.POST(
      '/question-banks/{bank_id}/questions/{question_id}/versions',
      {
        params: { path: { bank_id: bankId, question_id: questionId } },
        body: req,
      }
    )
  );
}

export async function publishQuestionVersion(
  bankId: string,
  questionId: string,
  versionId: string
): Promise<PublishQuestionVersionResult> {
  const client = await getOpenAPIClient();
  return unwrapData<PublishQuestionVersionResult>(
    await client.POST(
      '/question-banks/{bank_id}/questions/{question_id}/versions/{version_id}/publish',
      {
        params: {
          path: {
            bank_id: bankId,
            question_id: questionId,
            version_id: versionId,
          },
        },
      }
    )
  );
}

export { ApiResponseError };
