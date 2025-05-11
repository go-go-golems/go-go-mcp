import React from 'react';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as Yup from 'yup';
import { SearchParams } from '../types/scholarly';

interface SearchFormProps {
  onSearch: (params: SearchParams) => void;
  isLoading: boolean;
  availableSources: string[];
}

const SearchSchema = Yup.object().shape({
  query: Yup.string(),
  author: Yup.string(),
  title: Yup.string(),
  category: Yup.string(),
  'work-type': Yup.string(),
  // At least one search field is required
}).test(
  'atLeastOneField',
  'At least one search field (query, author, title) is required',
  value => Boolean(value.query || value.author || value.title || value.category || value['work-type'])
);

const SearchForm: React.FC<SearchFormProps> = ({ onSearch, isLoading, availableSources }) => {
  const initialValues: SearchParams = {
    query: '',
    sources: 'all',
    limit: 20,
    author: '',
    title: '',
    category: '',
    'work-type': '',
    'from-year': undefined,
    'to-year': undefined,
    sort: 'relevance',
    'open-access': '',
    'disable-rerank': false,
  };

  return (
    <div className="card mb-4 shadow-sm">
      <div className="card-header bg-primary text-white">
        <h4 className="mb-0">Scholarly Search</h4>
      </div>
      <div className="card-body">
        <Formik
          initialValues={initialValues}
          validationSchema={SearchSchema}
          onSubmit={(values) => {
            onSearch(values);
          }}
        >
          {({ errors, touched }) => (
            <Form>
              <div className="row g-3">
                <div className="col-md-12 mb-3">
                  <label htmlFor="query" className="form-label">Search Query</label>
                  <Field
                    type="text"
                    name="query"
                    className="form-control"
                    placeholder="Enter search terms"
                    disabled={isLoading}
                  />
                  <ErrorMessage name="query" component="div" className="text-danger" />
                </div>

                <div className="col-md-6 mb-3">
                  <label htmlFor="author" className="form-label">Author</label>
                  <Field
                    type="text"
                    name="author"
                    className="form-control"
                    placeholder="Author name"
                    disabled={isLoading}
                  />
                </div>

                <div className="col-md-6 mb-3">
                  <label htmlFor="title" className="form-label">Title</label>
                  <Field
                    type="text"
                    name="title"
                    className="form-control"
                    placeholder="Words in title"
                    disabled={isLoading}
                  />
                </div>

                <div className="col-md-6 mb-3">
                  <label htmlFor="category" className="form-label">Category (e.g., cs.AI)</label>
                  <Field
                    type="text"
                    name="category"
                    className="form-control"
                    placeholder="ArXiv category"
                    disabled={isLoading}
                  />
                </div>

                <div className="col-md-6 mb-3">
                  <label htmlFor="work-type" className="form-label">Work Type</label>
                  <Field
                    type="text"
                    name="work-type"
                    className="form-control"
                    placeholder="e.g., journal-article"
                    disabled={isLoading}
                  />
                </div>

                <div className="col-md-4 mb-3">
                  <label htmlFor="sources" className="form-label">Sources</label>
                  <Field
                    as="select"
                    name="sources"
                    className="form-select"
                    disabled={isLoading}
                  >
                    {availableSources.map(source => (
                      <option key={source} value={source}>{source}</option>
                    ))}
                  </Field>
                </div>

                <div className="col-md-4 mb-3">
                  <label htmlFor="limit" className="form-label">Result Limit</label>
                  <Field
                    type="number"
                    name="limit"
                    className="form-control"
                    min="1"
                    max="100"
                    disabled={isLoading}
                  />
                </div>

                <div className="col-md-4 mb-3">
                  <label htmlFor="sort" className="form-label">Sort Order</label>
                  <Field
                    as="select"
                    name="sort"
                    className="form-select"
                    disabled={isLoading}
                  >
                    <option value="relevance">Relevance</option>
                    <option value="newest">Newest</option>
                    <option value="oldest">Oldest</option>
                  </Field>
                </div>

                <div className="col-md-6 mb-3">
                  <label htmlFor="from-year" className="form-label">From Year</label>
                  <Field
                    type="number"
                    name="from-year"
                    className="form-control"
                    placeholder="Starting year"
                    disabled={isLoading}
                  />
                </div>

                <div className="col-md-6 mb-3">
                  <label htmlFor="to-year" className="form-label">To Year</label>
                  <Field
                    type="number"
                    name="to-year"
                    className="form-control"
                    placeholder="Ending year"
                    disabled={isLoading}
                  />
                </div>

                <div className="col-md-6 mb-3">
                  <label htmlFor="open-access" className="form-label">Open Access</label>
                  <Field
                    as="select"
                    name="open-access"
                    className="form-select"
                    disabled={isLoading}
                  >
                    <option value="">Any</option>
                    <option value="true">Yes</option>
                    <option value="false">No</option>
                  </Field>
                </div>

                <div className="col-md-6 mb-3 d-flex align-items-end">
                  <div className="form-check">
                    <Field
                      type="checkbox"
                      name="disable-rerank"
                      className="form-check-input"
                      id="disable-rerank"
                      disabled={isLoading}
                    />
                    <label className="form-check-label" htmlFor="disable-rerank">
                      Disable Reranking
                    </label>
                  </div>
                </div>

                {errors.query && touched.query && (
                  <div className="col-12">
                    <div className="alert alert-danger">
                      {errors.query}
                    </div>
                  </div>
                )}

                <div className="col-12 mt-4">
                  <button
                    type="submit"
                    className="btn btn-primary"
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <>
                        <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                        Searching...
                      </>
                    ) : 'Search Papers'}
                  </button>
                </div>
              </div>
            </Form>
          )}
        </Formik>
      </div>
    </div>
  );
};

export default SearchForm;