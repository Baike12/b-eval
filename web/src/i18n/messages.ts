import type { en } from './locales/en'

// Widen turns the literal types of the reference catalogue into plain
// strings.
//
// A translation may word a message differently from the reference, and the
// type of the catalogue must allow that while still fixing the set of keys.
type Widen<T> = { [K in keyof T]: T[K] extends string ? string : Widen<T[K]> }

// Messages is the shape every locale must have.
//
// The English catalogue is the reference. A locale that is missing a key is
// a compile error rather than a blank label in front of a reader.
export type Messages = Widen<typeof en>
